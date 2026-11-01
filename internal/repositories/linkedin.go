package repositories

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	customerrors "github.com/mvgoliveira/FetchJob-API/internal/errors"
	"github.com/mvgoliveira/FetchJob-API/internal/models"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

const (
	linkedInSearchURL = "https://www.linkedin.com/jobs-guest/jobs/api/seeMoreJobPostings/search"
	linkedInDetailURL = "https://www.linkedin.com/jobs-guest/jobs/api/jobPosting/%s"
	defaultUserAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
	maxShortDescRunes = 350
	detailConcurrency = 4
)

type LinkedInRepository interface {
	SearchJobs(ctx context.Context, params models.LinkedInSearchParams) ([]models.Job, error)
}

type LinkedInRepositoryImpl struct {
	httpClient *http.Client
}

func NewLinkedInRepository() LinkedInRepository {
	return &LinkedInRepositoryImpl{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type rawLinkedInCard struct {
	jobIDStr string
	job      models.Job
}

func (r *LinkedInRepositoryImpl) SearchJobs(ctx context.Context, params models.LinkedInSearchParams) ([]models.Job, error) {
	reqURL, err := buildSearchURL(params)
	if err != nil {
		log.Error().Str("method", "LinkedInRepository.SearchJobs.buildSearchURL").Err(err).Send()
		return nil, customerrors.ErrFetchingLinkedInJobs
	}

	doc, statusCode, err := r.fetchHTML(ctx, reqURL)
	if err != nil || statusCode != http.StatusOK {
		log.Error().
			Str("method", "LinkedInRepository.SearchJobs.fetchHTML").
			Int("status_code", statusCode).
			Err(err).
			Send()
		return nil, customerrors.ErrFetchingLinkedInJobs
	}

	cards := parseSearchCards(doc, params)
	if len(cards) == 0 {
		return []models.Job{}, nil
	}

	if params.Limit > 0 && len(cards) > params.Limit {
		cards = cards[:params.Limit]
	}

	jobs := make([]models.Job, len(cards))
	sem := make(chan struct{}, detailConcurrency)
	var wg sync.WaitGroup

	for i, card := range cards {
		wg.Add(1)
		go func(idx int, c rawLinkedInCard) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			job := c.job
			if c.jobIDStr != "" {
				r.enrichJobDetails(ctx, c.jobIDStr, &job, params)
			} else {
				job.Tags = extractTechTags(job.Name + " " + job.ShortDescription)
			}
			jobs[idx] = job
		}(i, card)
	}

	wg.Wait()

	return jobs, nil
}

func buildSearchURL(params models.LinkedInSearchParams) (string, error) {
	u, err := url.Parse(linkedInSearchURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("keywords", strings.TrimSpace(params.Keywords))

	if loc := strings.TrimSpace(params.Location); loc != "" {
		q.Set("location", loc)
	}

	workModel := strings.ToLower(strings.TrimSpace(params.WorkModel))
	if params.IsRemote != nil && *params.IsRemote {
		q.Set("f_WT", "2")
	} else {
		switch workModel {
		case string(models.WorkModelRemote):
			q.Set("f_WT", "2")
		case string(models.WorkModelHybrid):
			q.Set("f_WT", "3")
		case string(models.WorkModelOnsite):
			q.Set("f_WT", "1")
		}
	}

	if params.Start > 0 {
		q.Set("start", strconv.Itoa(params.Start))
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (r *LinkedInRepositoryImpl) fetchHTML(ctx context.Context, targetURL string) (*html.Node, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return doc, resp.StatusCode, nil
}

func parseSearchCards(doc *html.Node, params models.LinkedInSearchParams) []rawLinkedInCard {
	cardNodes := findAllNodes(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && hasClass(n, "base-search-card")
	})

	now := time.Now()
	results := make([]rawLinkedInCard, 0, len(cardNodes))

	for _, cardNode := range cardNodes {
		urn := getAttr(cardNode, "data-entity-urn")
		jobIDStr := extractJobIDFromURN(urn)
		var id uint
		if parsedID, err := strconv.ParseUint(jobIDStr, 10, 64); err == nil {
			id = uint(parsedID)
		}

		titleNode := findNode(cardNode, func(n *html.Node) bool {
			return n.Type == html.ElementNode && hasClass(n, "base-search-card__title")
		})
		companyNode := findNode(cardNode, func(n *html.Node) bool {
			return n.Type == html.ElementNode && hasClass(n, "base-search-card__subtitle")
		})
		locationNode := findNode(cardNode, func(n *html.Node) bool {
			return n.Type == html.ElementNode && hasClass(n, "job-search-card__location")
		})
		linkNode := findNode(cardNode, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "a" && hasClass(n, "base-card__full-link")
		})
		salaryNode := findNode(cardNode, func(n *html.Node) bool {
			return n.Type == html.ElementNode && hasClass(n, "job-search-card__salary-info")
		})
		timeNode := findNode(cardNode, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "time"
		})

		name := extractText(titleNode)
		company := extractText(companyNode)
		location := extractText(locationNode)
		jobURL := cleanLinkedInURL(getAttr(linkNode, "href"))

		if jobIDStr == "" && jobURL != "" {
			jobIDStr = extractJobIDFromURL(jobURL)
			if parsedID, err := strconv.ParseUint(jobIDStr, 10, 64); err == nil {
				id = uint(parsedID)
			}
		}

		if name == "" && jobURL == "" {
			continue
		}

		var salaryRange *string
		if s := extractText(salaryNode); s != "" {
			salaryRange = &s
		}

		createdAt := now
		if dt := getAttr(timeNode, "datetime"); dt != "" {
			if parsedTime, err := time.Parse("2006-01-02", dt); err == nil {
				createdAt = parsedTime
			}
		}

		workModel, isRemote := inferWorkModel(name+" "+location, params)

		results = append(results, rawLinkedInCard{
			jobIDStr: jobIDStr,
			job: models.Job{
				ID:               id,
				Name:             name,
				Company:          company,
				Url:              jobURL,
				SalaryRange:      salaryRange,
				ShortDescription: "",
				Location:         location,
				WorkModel:        workModel,
				IsRemote:         isRemote,
				Tags:             []string{},
				CreatedAt:        createdAt,
				UpdatedAt:        now,
			},
		})
	}

	return results
}

func (r *LinkedInRepositoryImpl) enrichJobDetails(ctx context.Context, jobID string, job *models.Job, params models.LinkedInSearchParams) {
	detailURL := fmt.Sprintf(linkedInDetailURL, jobID)
	doc, statusCode, err := r.fetchHTML(ctx, detailURL)
	if err != nil || statusCode != http.StatusOK {
		job.Tags = extractTechTags(job.Name)
		return
	}

	descNode := findNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && hasClass(n, "show-more-less-html__markup")
	})
	fullDesc := extractText(descNode)
	job.ShortDescription = truncateRunes(fullDesc, maxShortDescRunes)

	if job.SalaryRange == nil {
		salaryNode := findNode(doc, func(n *html.Node) bool {
			return n.Type == html.ElementNode && (hasClass(n, "compensation__salary") || hasClass(n, "salary"))
		})
		if s := extractText(salaryNode); s != "" {
			job.SalaryRange = &s
		}
	}

	criteriaItems := findAllNodes(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && hasClass(n, "description__job-criteria-item")
	})

	for _, item := range criteriaItems {
		headerNode := findNode(item, func(n *html.Node) bool {
			return n.Type == html.ElementNode && hasClass(n, "description__job-criteria-subheader")
		})
		valueNode := findNode(item, func(n *html.Node) bool {
			return n.Type == html.ElementNode && hasClass(n, "description__job-criteria-text")
		})

		header := strings.ToLower(extractText(headerNode))
		value := extractText(valueNode)
		if value == "" {
			continue
		}

		if strings.Contains(header, "seniority") || strings.Contains(header, "experiência") {
			valCopy := value
			job.Seniority = &valCopy
		}
	}

	workModel, isRemote := inferWorkModel(job.Name+" "+job.Location+" "+fullDesc, params)
	job.WorkModel = workModel
	job.IsRemote = isRemote
	job.Tags = extractTechTags(job.Name + " " + fullDesc)
}

func inferWorkModel(text string, params models.LinkedInSearchParams) (models.WorkModel, bool) {
	if params.IsRemote != nil && *params.IsRemote {
		return models.WorkModelRemote, true
	}

	switch strings.ToLower(strings.TrimSpace(params.WorkModel)) {
	case string(models.WorkModelRemote):
		return models.WorkModelRemote, true
	case string(models.WorkModelHybrid):
		return models.WorkModelHybrid, false
	case string(models.WorkModelOnsite):
		return models.WorkModelOnsite, false
	}

	lower := strings.ToLower(text)
	if strings.Contains(lower, "remote") || strings.Contains(lower, "remoto") || strings.Contains(lower, "home office") {
		return models.WorkModelRemote, true
	}
	if strings.Contains(lower, "hybrid") || strings.Contains(lower, "híbrido") || strings.Contains(lower, "hibrido") {
		return models.WorkModelHybrid, false
	}

	return models.WorkModelOnsite, false
}

var techKeywords = []struct {
	label   string
	pattern *regexp.Regexp
}{
	{"Go", regexp.MustCompile(`(?i)\b(golang|go)\b`)},
	{"Python", regexp.MustCompile(`(?i)\bpython\b`)},
	{"JavaScript", regexp.MustCompile(`(?i)\bjavascript\b`)},
	{"TypeScript", regexp.MustCompile(`(?i)\btypescript\b`)},
	{"Node.js", regexp.MustCompile(`(?i)\b(node\.js|nodejs)\b`)},
	{"React", regexp.MustCompile(`(?i)\b(react|reactjs|react\.js)\b`)},
	{"Next.js", regexp.MustCompile(`(?i)\b(next\.js|nextjs)\b`)},
	{"Vue", regexp.MustCompile(`(?i)\b(vue|vuejs|vue\.js)\b`)},
	{"Angular", regexp.MustCompile(`(?i)\bangular\b`)},
	{"Java", regexp.MustCompile(`(?i)\bjava\b`)},
	{"Spring Boot", regexp.MustCompile(`(?i)\bspring\s*boot\b`)},
	{"C#", regexp.MustCompile(`(?i)(\bc#|csharp|\.net\b)`)},
	{"PHP", regexp.MustCompile(`(?i)\bphp\b`)},
	{"Ruby", regexp.MustCompile(`(?i)\bruby\b`)},
	{"Rust", regexp.MustCompile(`(?i)\brust\b`)},
	{"Kotlin", regexp.MustCompile(`(?i)\bkotlin\b`)},
	{"Swift", regexp.MustCompile(`(?i)\bswift\b`)},
	{"PostgreSQL", regexp.MustCompile(`(?i)\b(postgresql|postgres)\b`)},
	{"MySQL", regexp.MustCompile(`(?i)\bmysql\b`)},
	{"MongoDB", regexp.MustCompile(`(?i)\bmongodb\b`)},
	{"Redis", regexp.MustCompile(`(?i)\bredis\b`)},
	{"SQL", regexp.MustCompile(`(?i)\bsql\b`)},
	{"Docker", regexp.MustCompile(`(?i)\bdocker\b`)},
	{"Kubernetes", regexp.MustCompile(`(?i)\b(kubernetes|k8s)\b`)},
	{"AWS", regexp.MustCompile(`(?i)\b(aws|amazon web services)\b`)},
	{"GCP", regexp.MustCompile(`(?i)\b(gcp|google cloud)\b`)},
	{"Azure", regexp.MustCompile(`(?i)\bazure\b`)},
	{"Terraform", regexp.MustCompile(`(?i)\bterraform\b`)},
	{"GraphQL", regexp.MustCompile(`(?i)\bgraphql\b`)},
	{"gRPC", regexp.MustCompile(`(?i)\bgrpc\b`)},
	{"Kafka", regexp.MustCompile(`(?i)\bkafka\b`)},
	{"RabbitMQ", regexp.MustCompile(`(?i)\brabbitmq\b`)},
}

func extractTechTags(text string) []string {
	tags := make([]string, 0)
	for _, tk := range techKeywords {
		if tk.pattern.MatchString(text) {
			tags = append(tags, tk.label)
		}
	}
	return tags
}

func extractJobIDFromURN(urn string) string {
	// Expected format: urn:li:jobPosting:1234567890
	parts := strings.Split(urn, ":")
	if len(parts) > 0 {
		last := strings.TrimSpace(parts[len(parts)-1])
		if _, err := strconv.ParseUint(last, 10, 64); err == nil {
			return last
		}
	}
	return ""
}

var urlIDRegex = regexp.MustCompile(`-(\d+)(?:\?|$|/)`)

func extractJobIDFromURL(rawURL string) string {
	matches := urlIDRegex.FindStringSubmatch(rawURL)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func cleanLinkedInURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if idx := strings.Index(rawURL, "?"); idx != -1 {
		return rawURL[:idx]
	}
	return rawURL
}

func truncateRunes(s string, maxRunes int) string {
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	runes := []rune(s)
	return strings.TrimSpace(string(runes[:maxRunes])) + "..."
}

func hasClass(n *html.Node, className string) bool {
	classes := strings.Fields(getAttr(n, "class"))
	for _, c := range classes {
		if c == className {
			return true
		}
	}
	return false
}

func getAttr(n *html.Node, key string) string {
	if n == nil {
		return ""
	}
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func extractText(n *html.Node) string {
	if n == nil {
		return ""
	}
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			sb.WriteString(node.Data)
			sb.WriteByte(' ')
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(sb.String()), " ")
}

func findNode(root *html.Node, match func(*html.Node) bool) *html.Node {
	if root == nil {
		return nil
	}
	if match(root) {
		return root
	}
	for c := root.FirstChild; c != nil; c = c.NextSibling {
		if found := findNode(c, match); found != nil {
			return found
		}
	}
	return nil
}

func findAllNodes(root *html.Node, match func(*html.Node) bool) []*html.Node {
	var results []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil {
			return
		}
		if match(n) {
			results = append(results, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	return results
}

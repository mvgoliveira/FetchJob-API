package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequesterID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requesterID := c.GetHeader("X-Requester-ID")
		if requesterID == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "X-Requester-ID header is required",
			})
			return
		}

		c.Set("requester_id", requesterID)
		c.Next()
	}
}

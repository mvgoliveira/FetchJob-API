package errors

import "errors"

var (
	ErrSendingTelegramMessage         = errors.New("could not send the telegram message")
	ErrSendingSlackMessage            = errors.New("could not send the slack message")
	ErrSlackTokenNotFound             = errors.New("could not resolve the slack bot token for the organization")
	ErrSendingWhatsAppMessage         = errors.New("could not send the whatsapp message")
	ErrSendingEmail                   = errors.New("could not send the email")
	ErrGettingRecipients              = errors.New("could not get the recipients")
	ErrMarshallingPrismaCredentials   = errors.New("could not marshaling the prisma credentials")
	ErrMarshallingWorkOrderJson       = errors.New("could not marshaling the work order json")
	ErrAuthenticatingInPrisma         = errors.New("could not authenticate in prisma")
	ErrDecodingPrismaAuthResponse     = errors.New("could not decode authentication response")
	ErrCreatingPrismaWorkOrder        = errors.New("could not create a prisma work order struct")
	ErrCreatingPrismaWorkOrderRequest = errors.New("could not create a prisma work order request")
	ErrSendingPrismaWorkOrderRequest  = errors.New("could not send the prisma work order request")
	ErrReadingPrismaWorkOrderResponse = errors.New("could not read the prisma work order response")
	ErrFetchingLinkedInJobs           = errors.New("could not fetch jobs from linkedin")
	ErrParsingLinkedInJobs            = errors.New("could not parse jobs html from linkedin")
)


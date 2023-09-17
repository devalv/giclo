package errors

const (
	ConfigError                = "failed to read config: `%v`"
	ConfigCheckError           = "config has bad value: `%v`"
	CreateDirectoryError       = "failed to create repos directory: `%v`"
	APIRequestCreateError      = "error making http request: `%v`"
	APIRequestSendError        = "error sending http request: `%v`"
	APIResponseBadStatusError  = "bad response status code"
	APIResponseTotalPagesError = "could not extract total pages"
	ReadResponseBodyError      = "could not read response body: `%v`"
	ResponseUnmarshalError     = "could not unmarshal JSON: `%v`"
	APILikedResponseError      = "could not get liked repos: `%v`"
)

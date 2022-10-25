package httpclient

import "fmt"

const (
	InvalidAuthenticationTokenCode = "InvalidAuthenticationToken"
)

type ResponseError struct {
	Code string
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("response error '%s'", e.Code)
}

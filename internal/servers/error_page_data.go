package servers

import (
	"fmt"
	"net/http"
)

type ErrorPageData struct {
	HostURL string

	ErrorMessage string
}

func NewErrorPageData(hostURL string, errorCode int) ErrorPageData {
	return ErrorPageData{
		HostURL:      hostURL,
		ErrorMessage: fmt.Sprintf("Error %d. %s", errorCode, http.StatusText(errorCode)),
	}
}

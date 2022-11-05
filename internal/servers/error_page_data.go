package servers

import (
	"fmt"
	"net/http"
)

type ErrorPageData struct {
	RedirectURI  string
	ErrorMessage string
}

func NewErrorPageData(redirectURI string, errorCode int) ErrorPageData {
	pageData := ErrorPageData{RedirectURI: redirectURI}

	pageData.ErrorMessage = fmt.Sprintf("Error %d. %s", errorCode, http.StatusText(errorCode))

	return pageData
}

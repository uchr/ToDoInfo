package servers

import (
	"fmt"
	"net/http"
)

type ErrorPageData struct {
	ErrorMessage string
}

func NewErrorPageData(errorCode int) ErrorPageData {
	return ErrorPageData{ErrorMessage: fmt.Sprintf("Error %d. %s", errorCode, http.StatusText(errorCode))}
}

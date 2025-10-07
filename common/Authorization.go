package common

import (
	"net/http"
)

func CheckKey(request *http.Request, key string) bool {
	requestKey := request.Header.Get("Key")
	if key == requestKey {
		return true
	} else {
		return false
	}
}
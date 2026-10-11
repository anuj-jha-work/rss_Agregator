package auth

import (
	"errors"
	"net/http"
	"strings"
)

// GetApikey extracts the api key from the headers of an HTTP request.
// Example usage:
// Authorization:Apikey {insert your api key here}
func GetApikey(headers http.Header) (string, error) {
	val := headers.Get("Authorization")
	if val == "" {
		return "", errors.New("no info is provided in the header")
	}
	vals := strings.Split(val, " ")
	if len(vals) != 2 {
		return "", errors.New("malformed header Is Given")
	}
	if vals[0] != "Apikey" {
		return "", errors.New("malformed Fisrt part of AUth Header")
	}
	return vals[1], nil
}

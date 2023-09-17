package mutator

import (
	"encoding/base64"
	"net/url"
)

type applyFn func(input string) string

var applyFunctions = map[string]applyFn{
	"url":    urlEncode,
	"base64": base64Encode,
}

func urlEncode(input string) string {
	return url.QueryEscape(input)
}

func base64Encode(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

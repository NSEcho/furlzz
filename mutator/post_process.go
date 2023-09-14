package mutator

import (
	"encoding/base64"
	"net/url"
)

type applyFn func(input []byte) []byte

var applyFunctions = map[string]applyFn{
	"url":    urlEncode,
	"base64": base64Encode,
}

func urlEncode(input []byte) []byte {
	return []byte(url.QueryEscape(string(input)))
}

func base64Encode(input []byte) []byte {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(input)))
	base64.StdEncoding.Encode(dst, input)
	return dst
}

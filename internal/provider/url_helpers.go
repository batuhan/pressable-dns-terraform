package provider

import (
	"net/url"
	"strconv"
	"strings"
)

func urlQueryEscape(value string) string {
	return url.QueryEscape(value)
}

func fmtInt(value int64) string {
	return strconv.FormatInt(value, 10)
}

func joinQuery(parts []string) string {
	return strings.Join(parts, "&")
}

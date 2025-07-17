package utils

import (
	"fmt"
	"strings"
)

func BuildUrlForAPIHost(host, path string) string {

	if strings.Contains(host, "localhost") || strings.Contains(host, "192.168") {
		return fmt.Sprintf("http://%s/%s", host, path)
	}
	return fmt.Sprintf("https://api.myproject.com/%s", path)
}

func BuildUrlForHost(host, path string) string {

	return fmt.Sprintf("https://%s/%s", host, path)
}

package utils

import "net/url"

func IsURL(u string) bool {
	_, err := url.ParseRequestURI(u)
	return err == nil
}

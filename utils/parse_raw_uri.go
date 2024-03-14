package utils

import (
	"net"
	"net/url"
	"strconv"
)

func ParseRawURI(uri string) (*url.URL, error) {
	if host, port, err := net.SplitHostPort(uri); err == nil {
		if _, err := strconv.Atoi(port); err == nil {
			return &url.URL{
				Host: host + ":" + port,
			}, nil
		}
	}
	return url.ParseRequestURI(uri)
}

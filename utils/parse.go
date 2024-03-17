package utils

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

var (
	ErrExecutionEndpointToManyPartsInID = errors.New(
		"too many parts in the id of execution endpoint",
	)
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

func ParseELEndpointID(id string) (
	group, name string, err error,
) {
	name = id

	if strings.Contains(id, ":") {
		parts := strings.Split(id, ":")
		if len(parts) > 2 {
			return "", "", fmt.Errorf("%w: %s",
				ErrExecutionEndpointToManyPartsInID, id,
			)
		}
		group = parts[0]
		name = parts[1]
	}

	return group, name, nil
}

func MakeELEndpointID(group, name string) string {
	if group == "" {
		return name
	}
	return group + ":" + name
}

package state

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrExecutionEndpointToManyPartsInID = errors.New("too many parts in the id of execution endpoint")
)

func parseExecutionPointID(id string) (
	name, namespace string, err error,
) {
	name = id

	if strings.Contains(id, ":") {
		parts := strings.Split(id, ":")
		if len(parts) > 2 {
			return "", "", fmt.Errorf("%w: %s",
				ErrExecutionEndpointToManyPartsInID, id,
			)
		}
		namespace = parts[0]
		name = parts[1]
	}

	return name, namespace, nil
}

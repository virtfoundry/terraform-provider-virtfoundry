package virtfoundry

import (
	"fmt"
	"strings"
)

func findByID[T any](items []T, id string, getID func(T) string) (*T, error) {
	for i := range items {
		if getID(items[i]) == id {
			return &items[i], nil
		}
	}
	return nil, fmt.Errorf("resource %q not found", id)
}

func findByIDOrName[T any](items []T, idOrName string, getID, getName func(T) string) (*T, error) {
	for i := range items {
		if getID(items[i]) == idOrName || getName(items[i]) == idOrName {
			return &items[i], nil
		}
	}
	return nil, fmt.Errorf("resource %q not found", idOrName)
}

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "HTTP 404") || strings.Contains(msg, "not found")
}

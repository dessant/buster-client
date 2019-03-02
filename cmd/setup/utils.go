package main

import (
	"os"

	"github.com/gofrs/uuid"
)

func stringInSlice(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func newToken() string {
	token, _ := uuid.NewV4()
	return token.String()
}

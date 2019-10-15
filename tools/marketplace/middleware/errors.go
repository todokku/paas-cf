package middleware

import "fmt"

type NotFoundError struct {
	Path string
}

func (n NotFoundError) Error() string {
	return fmt.Sprintf("URL Not found: %s", n.Path)
}

func ErrNotFound(path string) NotFoundError {
	return NotFoundError{Path: path}
}

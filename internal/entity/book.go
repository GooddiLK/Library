package entity

import "github.com/pkg/errors"

type Book struct {
	ID        string
	Name      string
	AuthorIDs []string
}

var (
	ErrBookNotFound      = errors.New("book not found")
	ErrBookAlreadyExists = errors.New("book already exists")
)

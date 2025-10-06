package entity

import (
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Book struct {
	Id        string
	Name      string
	AuthorIds []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	ErrBookNotFound      = status.Error(codes.NotFound, "book not found")
	ErrBookAlreadyExists = status.Error(codes.AlreadyExists, "book already exists")
)

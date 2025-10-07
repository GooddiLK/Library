package entity

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Author struct {
	Id   string
	Name string
}

var (
	ErrAuthorNotFound      = status.Error(codes.NotFound, "author not found")
	ErrAuthorAlreadyExists = status.Error(codes.AlreadyExists, "author already exists")
)

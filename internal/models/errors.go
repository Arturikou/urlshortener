package models

import "errors"

var (
	ErrAliasAlreadyExists = errors.New("alias already exists")
	ErrURLAlreadyExists   = errors.New("url already exists")
	ErrNotFound           = errors.New("not found")
)

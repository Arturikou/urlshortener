package models

import "errors"

var (
	ErrAliasAlreadyExists = errors.New("alias already exists")
	ErrURLAlreadyShorted  = errors.New("url already shorted")
	ErrURLAlreadyExists   = errors.New("url already exists")
	ErrNotFound           = errors.New("not found")
	ErrURLDeleted         = errors.New("url deleted")
)

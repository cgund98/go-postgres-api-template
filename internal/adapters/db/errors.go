package db

import "errors"

// ErrNoDBContext is returned when a repository cannot extract the database context from the context
var ErrNoDBContext = errors.New("database context not found in context")

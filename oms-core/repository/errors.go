package repository

import "errors"

var (
	ErrOrderNotFound               = errors.New("order not found")
	ErrAssemblyApplicationNotFound = errors.New("assembly application not found")
	ErrInvalidInput                = errors.New("invalid input parameters")
	ErrDatabaseOperation           = errors.New("database operation failed")
)

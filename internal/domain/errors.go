package domain

import "errors"

var ErrInvalidInput = errors.New("invalid input data")
var ErrInternal = errors.New("internal server error")
var ErrNotFound = errors.New("nothing was found")

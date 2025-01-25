package http

import (
	"github.com/pkg/errors"
)

var ErrTooManyRequests = errors.New("too many requests")

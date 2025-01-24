package coinrepository

import "github.com/pkg/errors"

var (
	ErrBannedCoinNotFound = errors.New("banned coin not found")
	ErrCoinNotFound       = errors.New("coin not found")
)

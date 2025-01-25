package badger

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/pkg/errors"
)

func getValue(item *badger.Item) ([]byte, error) {
	var value []byte
	err := item.Value(func(v []byte) error {
		value = make([]byte, len(v))
		copy(value, v)
		return nil
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return value, nil
}

package badger

import (
	"github.com/biosvos/coin-cache-service/internal/pkg/keyvalue"
	"github.com/dgraph-io/badger/v4"
	"github.com/pkg/errors"
)

var _ keyvalue.Store = (*Store)(nil)

type Store struct {
	db *badger.DB
}

func NewStore(path string) (*Store, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) Create(key []byte, value []byte) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err == nil {
			return keyvalue.ErrKeyAlreadyExists
		}
		return txn.Set(key, value)
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *Store) List(prefix []byte) ([][]byte, error) {
	var ret [][]byte
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{}) //nolint:exhaustruct
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			value, err := getValue(item)
			if err != nil {
				return errors.WithStack(err)
			}
			ret = append(ret, value)
		}
		return nil
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ret, nil
}

func (s *Store) Get(key []byte) ([]byte, error) {
	var ret []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return errors.WithStack(err)
		}
		value, err := getValue(item)
		if err != nil {
			return errors.WithStack(err)
		}
		ret = value
		return nil
	})
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, keyvalue.ErrKeyNotFound
		}
		return nil, errors.WithStack(err)
	}
	return ret, nil
}

func (s *Store) Update(key []byte, value []byte) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err != nil {
			return errors.WithStack(err)
		}
		return txn.Set(key, value)
	})
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return keyvalue.ErrKeyNotFound
		}
		return errors.WithStack(err)
	}
	return nil
}

func (s *Store) Delete(key []byte) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err != nil {
			return errors.WithStack(err)
		}
		return txn.Delete(key)
	})
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return keyvalue.ErrKeyNotFound
		}
		return errors.WithStack(err)
	}
	return nil
}

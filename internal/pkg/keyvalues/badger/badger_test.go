package badger_test

import (
	"testing"

	"github.com/biosvos/coin-cache-service/internal/pkg/keyvalue"
	"github.com/biosvos/coin-cache-service/internal/pkg/keyvalues/badger"
	"github.com/stretchr/testify/require"
)

func TestBadger_Create(t *testing.T) {
	t.Parallel()
	store, err := badger.NewStore(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() {
		store.Close()
	})
	err = store.Create([]byte("key"), []byte("value"))
	require.NoError(t, err)

	err = store.Create([]byte("key"), []byte("value2"))

	require.ErrorIs(t, err, keyvalue.ErrKeyAlreadyExists)
}

func TestBadger_Get(t *testing.T) {
	t.Parallel()
	store, err := badger.NewStore(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() {
		store.Close()
	})

	value, err := store.Get([]byte("key"))

	require.ErrorIs(t, err, keyvalue.ErrKeyNotFound)
	require.Empty(t, value)
}

func TestBadger_Update(t *testing.T) {
	t.Parallel()
	store, err := badger.NewStore(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() {
		store.Close()
	})

	err = store.Update([]byte("key"), []byte("value"))

	require.ErrorIs(t, err, keyvalue.ErrKeyNotFound)
}

func TestBadger_Delete(t *testing.T) {
	t.Parallel()
	store, err := badger.NewStore(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() {
		store.Close()
	})

	err = store.Delete([]byte("key"))

	require.ErrorIs(t, err, keyvalue.ErrKeyNotFound)
}

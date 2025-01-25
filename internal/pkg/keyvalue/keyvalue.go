package keyvalue

type Store interface {
	Create(key []byte, value []byte) error
	List(prefix []byte) ([][]byte, error)
	Get(key []byte) ([]byte, error)
	Update(key []byte, value []byte) error
	Delete(key []byte) error
}

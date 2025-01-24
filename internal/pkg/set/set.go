package setpkg

type Set[K comparable, V any] struct {
	items map[K]V
	fn    func(V) K
}

func NewSet[K comparable, V any](fn func(V) K) *Set[K, V] {
	return &Set[K, V]{
		items: make(map[K]V),
		fn:    fn,
	}
}

func (s *Set[K, V]) Add(items ...V) {
	for _, item := range items {
		s.items[s.fn(item)] = item
	}
}

func (s *Set[K, V]) Delete(items ...V) {
	for _, item := range items {
		delete(s.items, s.fn(item))
	}
}

func (s *Set[K, V]) Contain(item V) bool {
	_, ok := s.items[s.fn(item)]
	return ok
}

func (s *Set[K, V]) ContainKey(key K) bool {
	_, ok := s.items[key]
	return ok
}

func (s *Set[K, V]) Intersection(other *Set[K, V]) *Set[K, V] {
	ret := NewSet(s.fn)
	for key, value := range s.items {
		if other.ContainKey(key) {
			ret.Add(value)
		}
	}
	return ret
}

func (s *Set[K, V]) Difference(other *Set[K, V]) *Set[K, V] {
	ret := NewSet(s.fn)
	for key, value := range s.items {
		if !other.ContainKey(key) {
			ret.Add(value)
		}
	}
	return ret
}

func (s *Set[K, V]) Values() []V {
	var ret []V
	for _, value := range s.items {
		ret = append(ret, value)
	}
	return ret
}

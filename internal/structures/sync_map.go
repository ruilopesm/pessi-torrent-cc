package structures

import "sync"

type SynchronizedMap[K comparable, V any] struct {
	M map[K]V
	sync.Mutex
}

func NewSynchronizedMap[K comparable, V any]() SynchronizedMap[K, V] {
	return SynchronizedMap[K, V]{
		M: make(map[K]V),
	}
}

func (m *SynchronizedMap[K, V]) Get(key K) (V, bool) {
	m.Lock()
	defer m.Unlock()

	val, ok := m.M[key]
	return val, ok
}

func (m *SynchronizedMap[K, V]) Put(key K, val V) {
	m.Lock()
	defer m.Unlock()

	m.M[key] = val
}

func (m *SynchronizedMap[K, V]) Delete(key K) {
	m.Lock()
	defer m.Unlock()

	delete(m.M, key)
}

func (m *SynchronizedMap[K, V]) Len() int {
	m.Lock()
	defer m.Unlock()

	return len(m.M)
}

func (m *SynchronizedMap[K, V]) Keys() []K {
	m.Lock()
	defer m.Unlock()

	keys := make([]K, 0, len(m.M))
	for k := range m.M {
		keys = append(keys, k)
	}

	return keys
}

func (m *SynchronizedMap[K, V]) Values() []V {
	m.Lock()
	defer m.Unlock()

	values := make([]V, 0, len(m.M))
	for _, v := range m.M {
		values = append(values, v)
	}

	return values
}

func (m *SynchronizedMap[K, V]) Contains(key K) bool {
	m.Lock()
	defer m.Unlock()

	_, ok := m.M[key]
	return ok
}

func (m *SynchronizedMap[K, V]) ForEach(f func(K, V)) {
	m.Lock()
	defer m.Unlock()

	for k, v := range m.M {
		f(k, v)
	}
}

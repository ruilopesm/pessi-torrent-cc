package structures

import "sync"

type SynchronizedMap[V any] struct {
	M map[string]V
	sync.Mutex
}

func NewSynchronizedMap[V any]() SynchronizedMap[V] {
	return SynchronizedMap[V]{M: make(map[string]V)}
}

func (m *SynchronizedMap[V]) Get(key string) (V, bool) {
	m.Lock()
	defer m.Unlock()

	val, ok := m.M[key]
	return val, ok
}

func (m *SynchronizedMap[V]) Put(key string, val V) {
	m.Lock()
	defer m.Unlock()

	m.M[key] = val
}

func (m *SynchronizedMap[V]) Delete(key string) {
	m.Lock()
	defer m.Unlock()

	delete(m.M, key)
}

func (m *SynchronizedMap[V]) Len() int {
	m.Lock()
	defer m.Unlock()

	return len(m.M)
}

func (m *SynchronizedMap[V]) Keys() []string {
	m.Lock()
	defer m.Unlock()

	keys := make([]string, 0, len(m.M))
	for k := range m.M {
		keys = append(keys, k)
	}
	return keys
}

func (m *SynchronizedMap[V]) Values() []V {
	m.Lock()
	defer m.Unlock()

	values := make([]V, 0, len(m.M))
	for _, v := range m.M {
		values = append(values, v)
	}
	return values
}

func (m *SynchronizedMap[V]) Contains(key string) bool {
	m.Lock()
	defer m.Unlock()

	_, ok := m.M[key]
	return ok
}

func (m *SynchronizedMap[V]) ForEach(f func(string, V)) {
	m.Lock()
	defer m.Unlock()

	for k, v := range m.M {
		f(k, v)
	}
}

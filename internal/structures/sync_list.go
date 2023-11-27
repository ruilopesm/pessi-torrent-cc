package structures

import (
	"errors"
	"sync"
)

type SynchronizedList[V comparable] struct {
	L []V
	sync.Mutex
}

func NewSynchronizedList[V comparable]() SynchronizedList[V] {
	return SynchronizedList[V]{L: make([]V, 0)}
}

func NewSynchronizedListWithInitialSize[V comparable](size uint) SynchronizedList[V] {
	return SynchronizedList[V]{L: make([]V, size)}
}

func (l *SynchronizedList[V]) Add(val V) {
	l.Lock()
	defer l.Unlock()

	l.L = append(l.L, val)
}

func (l *SynchronizedList[V]) Remove(val V) {
	l.Lock()
	defer l.Unlock()

	for i, v := range l.L {
		if v == val {
			l.L = append(l.L[:i], l.L[i+1:]...)
			break
		}
	}
}

func (l *SynchronizedList[V]) Set(index uint, val V) error {
	l.Lock()
	defer l.Unlock()

	if index >= uint(len(l.L)) {
		return errors.New("index out of bounds")
	}

	l.L[index] = val

	return nil
}

func (l *SynchronizedList[V]) Get(index uint) (V, error) {
	l.Lock()
	defer l.Unlock()

	if index >= uint(len(l.L)) {
		var zero V
		return zero, errors.New("index out of bounds")
	}

	return l.L[index], nil
}

func (l *SynchronizedList[V]) Len() int {
	l.Lock()
	defer l.Unlock()

	return len(l.L)
}

func (l *SynchronizedList[V]) Contains(val V) bool {
	l.Lock()
	defer l.Unlock()

	for _, v := range l.L {
		if v == val {
			return true
		}
	}

	return false
}

func (l *SynchronizedList[V]) ForEach(f func(V)) {
	l.Lock()
	defer l.Unlock()

	for _, v := range l.L {
		f(v)
	}
}

func (l *SynchronizedList[V]) Filter(predicate func(V) bool) []V {
	l.Lock()
	defer l.Unlock()

	var result []V
	for _, v := range l.L {
		if predicate(v) {
			result = append(result, v)
		}
	}

	return result
}

func (l *SynchronizedList[V]) IndexesWhere(predicate func(V) bool) []uint {
	l.Lock()
	defer l.Unlock()

	var result []uint
	for i, v := range l.L {
		if predicate(v) {
			result = append(result, uint(i))
		}
	}

	return result
}

package structures

import "sync"

type SynchronizedList[V comparable] struct {
	L []V
	sync.RWMutex
}

func NewSynchronizedList[V comparable](initialSize uint) SynchronizedList[V] {
	return SynchronizedList[V]{L: make([]V, 0, initialSize)}
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
			l.L[i] = l.L[len(l.L)-1]
			l.L = l.L[:len(l.L)-1]

			return
		}
	}
}

func (l *SynchronizedList[V]) Len() int {
	l.RLock()
	defer l.RUnlock()

	return len(l.L)
}

func (l *SynchronizedList[V]) Contains(val V) bool {
	l.RLock()
	defer l.RUnlock()

	for _, v := range l.L {
		if v == val {
			return true
		}
	}

	return false
}

func (l *SynchronizedList[V]) ForEach(f func(V)) {
	l.RLock()
	defer l.RUnlock()

	for _, v := range l.L {
		f(v)
	}
}

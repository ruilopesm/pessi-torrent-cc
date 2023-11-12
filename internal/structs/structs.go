package structs

import "sync"

type SynchronizedMap[T any] struct {
	M map[string]T
	sync.RWMutex
}

type SynchronizedList[T any] struct {
	L []T
	sync.RWMutex
}

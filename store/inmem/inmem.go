package inmem

import "sync"

type Inmem struct {
	store  map[string]any
	length int
	mu     sync.Mutex
}

func NewInmem() *Inmem {
	return &Inmem{
		store: make(map[string]any),
	}
}

func (i *Inmem) Insert(key string, value any) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.store[key] = value
}

func (i *Inmem) Find(key string) any {
	i.mu.Lock()
	defer i.mu.Unlock()
	v, ok := i.store[key]
	if !ok {
		return nil
	} else {
		return v
	}
}

func (i *Inmem) Len() int {
	return i.length
}

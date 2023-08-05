package store

type Store interface {
	Insert(key string, value any)
	Find(key string) any
}

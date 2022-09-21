package protocols

type Iterator[T any] interface {
	Next() (T, error)
	HasNext() bool
}

type Aggregate[T any] interface{ CreateIterator() Iterator[T] }

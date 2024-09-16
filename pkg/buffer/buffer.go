package buffer

type Buffer[T any] struct {
	buffer   []T
	position int
}

func New[T any](size int) *Buffer[T] {
	return &Buffer[T]{
		buffer:   make([]T, size),
		position: 0,
	}
}

func (b *Buffer[T]) Push(v T) bool {
	if b.position == len(b.buffer) {
		return true
	}
	b.buffer[b.position] = v
	b.position++
	return b.position == len(b.buffer)
}

func (b *Buffer[T]) Full() bool {
	return b.position == len(b.buffer)
}

func (b *Buffer[T]) Empty() bool {
	return b.position == 0
}

func (b *Buffer[T]) Position() int {
	return b.position
}

func (b *Buffer[T]) Cap() int {
	return cap(b.buffer)
}

func (b *Buffer[T]) Items() []T {
	return b.buffer[:b.position]
}

func (b *Buffer[T]) Clear() {
	b.position = 0
}

package geoindex

import (
	"bytes"
	"fmt"
)

type queue struct {
	elements []interface{}
	start    int64
	end      int64
	size     int
	cap      int
}

// NewQueue creates new Queue with initial capacity.
func newQueue(capacity int) *queue {
	return &queue{make([]interface{}, capacity, capacity), 0, 0, 0, capacity}
}

func (queue *queue) resize(size int) {
	newElements := make([]interface{}, size, size)

	for i := queue.start; i < queue.end; i++ {
		el := queue.elements[i%int64(queue.cap)]
		newElements[i-queue.start] = el
	}

	queue.cap = size
	queue.elements = newElements
	queue.start = 0
	queue.end = int64(queue.size)
}

// Push adds an element at the end of the queue.
func (queue *queue) Push(element interface{}) {
	if queue.size == queue.cap {
		queue.resize(queue.cap * 2)
	}

	queue.elements[queue.end%int64(queue.cap)] = element
	queue.end++
	queue.size++
}

// Pop removes an element from the front of the queue.
func (queue *queue) Pop() interface{} {
	if queue.size == 0 {
		return nil
	}

	if queue.size < queue.cap/4 && queue.size > 4 {
		queue.resize(queue.cap / 2)
	}

	result := queue.elements[queue.start%int64(queue.cap)]
	queue.start++
	queue.size--

	return result
}

// Peek returns the element at the front of the queue.
func (queue *queue) Peek() interface{} {
	if queue.size == 0 {
		return nil
	}
	return queue.elements[queue.start%int64(queue.cap)]
}

// PeekBack returns the element at the back of the queue.
func (queue *queue) PeekBack() interface{} {
	if queue.size == 0 {
		return nil
	}

	return queue.elements[(queue.end-1)%int64(queue.cap)]
}

// Size returns the number of elements in the queue.
func (queue *queue) Size() int {
	return queue.size
}

// IsEmpty returns true if the queue is empty.
func (queue *queue) IsEmpty() bool {
	return queue.size == 0
}

// ForEach calls process for each element in the queue.
func (queue *queue) ForEach(process func(interface{})) {
	for i := queue.start; i < queue.end; i++ {
		process(queue.elements[i%int64(queue.cap)])
	}
}

func (queue *queue) String() string {
	var buffer bytes.Buffer

	buffer.WriteString("[")
	first := true
	queue.ForEach(func(element interface{}) {
		if !first {
			buffer.WriteString(", ")
		}

		buffer.WriteString(fmt.Sprintf("%s", element))
		first = false
	})
	buffer.WriteString("]")

	return buffer.String()
}

//go:generate genny -in=queue_template.go -out=queue_generated.go gen "ValType=string"

// //go:generate genny -in=$GOFILE -out=$GOFILE gen "ValType=BUILTINS,*MyType"

// https://github.com/cheekybits/genny
// Нужно просто вызвать в данной папке "go generate", в корне "go generate ./queue"
// Документация по генерации "go help generate"

package queue

import "github.com/cheekybits/genny/generic"

// NOTE: this is how easy it is to define a generic type
type ValType generic.Type

// SomethingQueue is a queue of Somethings.
type ValTypeQueue struct {
	items []ValType
}

func NewValTypeQueue() *ValTypeQueue {
	return &ValTypeQueue{items: make([]ValType, 0)}
}
func (q *ValTypeQueue) Push(item ValType) {
	q.items = append(q.items, item)
}
func (q *ValTypeQueue) Pop() ValType {
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

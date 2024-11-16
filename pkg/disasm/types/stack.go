package types

type Stack[T any] struct {
	items []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{nil}
}

func (stack *Stack[T]) Push(item T) {
	stack.items = append(stack.items, item)
}

func (stack *Stack[T]) Pop() T {
	var result T
	if len(stack.items) == 0 {
		panic("No items available")
	}

	l := len(stack.items)

	result = stack.items[l-1]
	stack.items = stack.items[:l-1]
	return result
}

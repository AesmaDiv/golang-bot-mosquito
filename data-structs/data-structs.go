package datastructs

import "fmt"

type Node struct {
	Value any
	Next  *Node
}

type Stack struct {
	next *Node
}

func (s *Stack) Display() {
	index := 0
	node := s.next
	for node != nil {
		fmt.Printf("%d: %v\n", index, node.Value)
		node = node.Next
		index++
	}
}

func (s *Stack) Push(value any) {
	node := &Node{Value: value, Next: s.next}
	s.next = node
}
func (s *Stack) Pop() *Node {
	node := s.next
	if node != nil {
		s.next = node.Next
	}
	return node
}

type Queue struct {
	Stack
	last *Node
}

func (q *Queue) Push(value any) {
	node := &Node{Value: value}
	if q.next == nil {
		q.next = node
	} else if q.last != nil {
		q.last.Next = node
	}
	q.last = node
}
func (q *Queue) Pop() *Node {
	node := q.next
	if node != nil {
		q.next = q.next.Next
	}
	return node
}

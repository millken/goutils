package trie

import (
	"fmt"
	"strings"
	"sync"
)

func segmenter(path string, pos int) (segment string, next int) {
	if len(path) == 0 || pos < 0 || pos > len(path) {
		return "", -1
	}
	start := strings.LastIndexByte(path[:pos], '.')
	if start == -1 {
		return path[:pos], -1
	}
	return path[start+1 : pos], start
}

type DomainTrie[T any] struct {
	sync.RWMutex
	root *Node[T]
}

func NewDomainTrie[T any]() *DomainTrie[T] {
	return &DomainTrie[T]{root: NewNodeNil[T](nil)}
}

func (t *DomainTrie[T]) Insert(k string, value T) {
	t.Lock()
	defer t.Unlock()
	node := t.root
	i := len(k)
	var part string
	for i > 0 {
		part, i = segmenter(k, i)
		node = node.GetOrSet(part, NewNodeNil[T](node))
	}
	node.MarkAsLeaf()
	node.SetData(value)
}

func (t *DomainTrie[T]) Search(k string) *Node[T] {
	t.RLock()
	n := t.search(t.root, k)
	t.RUnlock()
	return n
}

func (t *DomainTrie[T]) search(node *Node[T], k string) *Node[T] {
	pos := len(k)
	for pos > 0 {
		segment, next := segmenter(k, pos)
		if nextNode, exists := node.children[segment]; exists {
			if next == -1 {
				return nextNode
			}
			if n := t.search(nextNode, k[:next]); n != nil && n.IsLeaf() {
				return n
			}
		}
		if nextNode, exists := node.children["*"]; exists {
			if next == -1 {
				return nextNode
			}
			if n := t.search(nextNode, k[:next]); n != nil && n.IsLeaf() {
				return n
			}
		}
		pos = next
	}
	return nil
}

func (t *DomainTrie[T]) print(n *Node[T], key string, space int) {
	if n == nil {
		return
	}
	space += 10

	n.ForEach(func(s string, nc *Node[T]) {
		t.print(nc, s, space)
	})

	for i := 0; i < space; i++ {
		fmt.Print(" ")
	}

	if key == "" {
		if n != t.root {
			fmt.Printf("+: %v\n", n.data)
		} else {
			fmt.Print("root\n")
		}
	} else {
		fmt.Printf("%s: %v\n", key, n.data)
	}

}

func (t *DomainTrie[T]) Print() {
	t.print(t.root, "", 0)
}

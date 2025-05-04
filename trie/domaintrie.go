package trie

import (
	"fmt"
	"iter"
	"sync"
)

func segmenter(path string, pos int) (segment string, next int) {
	if len(path) == 0 || pos < 0 || pos > len(path) {
		return "", -1
	}
	for i := pos - 1; i >= 0; i-- {
		if path[i] == '.' {
			return path[i+1 : pos], i
		}
	}
	return path[:pos], -1
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
		node = node.GetOrSet(part, NewNodeNil(node))
	}
	node.MarkAsLeaf()
	node.SetData(value)
}
func SplitDomainReverseIterator(domain string) iter.Seq[string] {
	return func(yield func(string) bool) {
		start := len(domain)
		// 从右向左扫描
		for i := len(domain) - 1; i >= 0; i-- {
			if domain[i] == '.' {
				part := domain[i+1 : start]
				if part != "" {
					if !yield(part) {
						return
					}
				}
				start = i // 更新段起始位置
			}
		}
		// 处理首段（最左侧部分）
		if start > 0 {
			part := domain[0:start]
			if !yield(part) {
				return
			}
		}
	}
}
func (tree *DomainTrie[T]) Insert2(domain string, data T) {
	tree.Lock()
	defer tree.Unlock()
	node := tree.root
	for part := range SplitDomainReverseIterator(domain) {
		child, ok := node.children[part]
		if !ok {
			child = &Node[T]{children: make(map[string]*Node[T])}
			node.children[part] = child
			child.parent = node
		}
		node = child
	}
	node.data = data
	node.isLeaf = true
}

func (tree *DomainTrie[T]) Search2(domain string) *Node[T] {
	tree.RLock()
	defer tree.RUnlock()

	var wildcard *Node[T]
	node := tree.root
	matchCount := 0 // 记录成功匹配的段数
	totalCount := 0 // 记录总段数
	hc := []byte{0, 0}

	for part := range SplitDomainReverseIterator(domain) {
		totalCount++
		child, ok := node.children[part]
		if !ok {
			// 精确匹配失败，尝试通配符
			child, ok = node.children["*"]
			if ok {
				wildcard = child
				matchCount++
				hc[1]++
				break
			}
		} else {
			node = child
			matchCount++
			hc[0]++
		}
	}

	// 只有完全匹配所有段时才算成功
	if matchCount == totalCount {
		// 通配符匹配
		if wildcard != nil && wildcard.isLeaf {
			return wildcard
		}
		if node.isLeaf {
			return node
		}

	}
	return nil
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
		if foundNode := t.findNextNode(node, segment, k, next); foundNode != nil {
			return foundNode
		}
		if foundNode := t.findNextNode(node, "*", k, next); foundNode != nil {
			return foundNode
		}
		pos = next
	}
	return nil
}

func (t *DomainTrie[T]) findNextNode(node *Node[T], segment, k string, next int) *Node[T] {
	if nextNode, exists := node.children[segment]; exists {
		if next == -1 {
			return nextNode
		}
		if n := t.search(nextNode, k[:next]); n != nil && n.IsLeaf() {
			return n
		}
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

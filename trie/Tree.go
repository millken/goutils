package trie

import (
	"fmt"
	"iter"
	"strings"
	"sync"
)

// DomainTree 表示域名树
type DomainTree[T any] struct {
	sync.RWMutex
	root *DomainNode[T]
}

func NewDomainTree[T any]() *DomainTree[T] {
	return &DomainTree[T]{root: newDomainNode[T]()}
}

// Add 添加域名到树中
func (tree *DomainTree[T]) Add(domain string, data T) {
	tree.Lock()
	defer tree.Unlock()

	tree.root.add(domain, data)
}

// Lookup 查找域名
func (tree *DomainTree[T]) Lookup(domain string) (T, bool) {
	tree.RLock()
	defer tree.RUnlock()
	return tree.root.lookup(domain)
}

func (tree *DomainTree[T]) Delete(domain string) {
	tree.Lock()
	defer tree.Unlock()

	tree.root.delete(domain)
}

// Map 遍历并转换所有节点数据
// func (tree *DomainTree[T]) Map(transform func(T) T) {
// 	tree.root.each(func(node *DomainNode[T]) {
// 		node.data = transform(node.data)
// 	})
// }

func (tree *DomainTree[T]) Print() {
	tree.root.each(func(deep int, label string, node *DomainNode[T]) {
		if node.isLeaf {
			fmt.Printf("%s", strings.Repeat(" ", deep*8))
		} else {
			fmt.Printf("%s", strings.Repeat(" ", deep*8))
		}
		fmt.Printf("%s|", label)
		fmt.Println()
	})
}

func splitDomainReverseIterator(domain string) iter.Seq[string] {
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

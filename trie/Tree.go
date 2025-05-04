package trie

import (
	"fmt"
	"iter"
	"sync"
)

// DomainTree 表示域名树
type DomainTree[T any] struct {
	sync.RWMutex
	root *DomainNode[T]
}

func NewDomainTree[T any]() *DomainTree[T] {
	return &DomainTree[T]{root: newDomainNode[T]("", 0)}
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
func (tree *DomainTree[T]) Map(transform func(T) T) {
	tree.Lock()
	defer tree.Unlock()
	tree.root.each(func(node *DomainNode[T]) {
		if node.isLeaf {
			node.data = transform(node.data)
		}
	})
}

// All 返回一个迭代器，遍历所有域名及其配置
func (tree *DomainTree[T]) All() iter.Seq[struct {
	Domain string
	Value  T
}] {
	return func(yield func(struct {
		Domain string
		Value  T
	}) bool) {
		var walk func(node *DomainNode[T], labels []string) bool
		walk = func(node *DomainNode[T], labels []string) bool {
			if node.isLeaf {
				// 域名应从 labels 反转拼接
				domain := ""
				for i := len(labels) - 1; i >= 0; i-- {
					if i != len(labels)-1 {
						domain += "."
					}
					domain += labels[i]
				}
				if !yield(struct {
					Domain string
					Value  T
				}{Domain: domain, Value: node.data}) {
					return false
				}
			}
			for _, child := range node.children {
				if !walk(child, append(labels, child.label)) {
					return false
				}
			}
			return true
		}
		walk(tree.root, nil)
	}
}

func (tree *DomainTree[T]) Reset() {
	tree.Lock()
	defer tree.Unlock()
	tree.root = newDomainNode[T]("", 0)
}

func (tree *DomainTree[T]) Print() {
	var printNode func(node *DomainNode[T], prefix string, isLast bool)
	printNode = func(node *DomainNode[T], prefix string, isLast bool) {
		if node.label != "" {
			branch := "├── "
			if isLast {
				branch = "└── "
			}
			leafMark := ""
			if node.isLeaf {
				leafMark = "."
			}
			fmt.Printf("%s%s%s%s\n", prefix, branch, node.label, leafMark)
			if isLast {
				prefix += "    "
			} else {
				prefix += "│   "
			}
		}
		for i, child := range node.children {
			printNode(child, prefix, i == len(node.children)-1)
		}
	}
	printNode(tree.root, "", true)
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

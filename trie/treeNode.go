package trie

// DomainNode 表示域名树节点
type DomainNode[T any] struct {
	children map[string]*DomainNode[T]
	data     T
	isLeaf   bool
}

func newDomainNode[T any]() *DomainNode[T] {
	var empty T
	return &DomainNode[T]{
		children: make(map[string]*DomainNode[T]),
		data:     empty,
		isLeaf:   false,
	}
}

// addChild 添加子节点
func (node *DomainNode[T]) add(domain string, data T) {
	if node.children == nil {
		node.children = make(map[string]*DomainNode[T])
	}

	// 从右向左遍历域名
	for part := range splitDomainReverseIterator(domain) {
		child, ok := node.children[part]
		if !ok {
			child = &DomainNode[T]{children: make(map[string]*DomainNode[T])}
			node.children[part] = child
		}
		node = child
	}
	node.data = data
	node.isLeaf = true

}

func (node *DomainNode[T]) delete(domain string) {
	if node == nil {
		return
	}
	type pathElem struct {
		parent *DomainNode[T]
		label  string
	}
	var path []pathElem
	curr := node
	for part := range splitDomainReverseIterator(domain) {
		child, ok := curr.children[part]
		if !ok {
			return
		}
		path = append(path, pathElem{curr, part})
		curr = child
	}

	if curr.isLeaf {
		curr.isLeaf = false
		var empty T
		curr.data = empty
	}
	for i := len(path) - 1; i >= 0; i-- {
		parent := path[i].parent
		label := path[i].label
		child := parent.children[label]
		if len(child.children) == 0 && !child.isLeaf {
			delete(parent.children, label)
		} else {
			break
		}
	}
}

func (node *DomainNode[T]) lookup(domain string) (T, bool) {
	var empty T
	// 从右向左遍历域名
	for part := range splitDomainReverseIterator(domain) {
		child, ok := node.children[part]
		if !ok {
			// 精确匹配失败，尝试通配符
			child, ok = node.children["*"]
			if ok && child.isLeaf {
				return child.data, true
			}
			return empty, false
		}
		node = child
	}
	if node.isLeaf {
		return node.data, true
	}
	return empty, false
}

// each 遍历树中所有节点
func (node *DomainNode[T]) each(callback func(int, string, *DomainNode[T])) {
	// callback(deep, label, child)
	// deep := 1
	// for label, child := range node.children {
	// 	if child != nil {
	// 		child.each(callback)
	// 	}
	// 	deep++
	// }
}

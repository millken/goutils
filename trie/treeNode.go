package trie

type DomainNode[T any] struct {
	children []*DomainNode[T]
	label    string
	data     T
	deep     int
	isLeaf   bool
}

// 新建节点
func newDomainNode[T any](label string, deep int) *DomainNode[T] {
	var empty T
	return &DomainNode[T]{
		children: nil,
		label:    label,
		data:     empty,
		deep:     deep,
		isLeaf:   false,
	}
}

// 查找子节点
func (node *DomainNode[T]) findChild(label string) (*DomainNode[T], int) {
	for i, child := range node.children {
		if child.label == label {
			return child, i
		}
	}
	return nil, -1
}

// 添加子节点
func (node *DomainNode[T]) add(domain string, data T) {
	curr := node
	deep := node.deep
	for part := range splitDomainReverseIterator(domain) {
		child, idx := curr.findChild(part)
		if idx == -1 {
			child = newDomainNode[T](part, deep+1)
			curr.children = append(curr.children, child)
		}
		curr = child
		deep++
	}
	curr.data = data
	curr.isLeaf = true
}

// 删除节点
func (node *DomainNode[T]) delete(domain string) {
	type pathElem struct {
		parent *DomainNode[T]
		label  string
		idx    int
	}
	var path []pathElem
	curr := node
	for part := range splitDomainReverseIterator(domain) {
		child, idx := curr.findChild(part)
		if child == nil {
			return
		}
		path = append(path, pathElem{curr, part, idx})
		curr = child
	}
	if curr.isLeaf {
		curr.isLeaf = false
		var empty T
		curr.data = empty
	}
	// 回溯清理无用节点
	for i := len(path) - 1; i >= 0; i-- {
		parent := path[i].parent
		idx := path[i].idx
		child := parent.children[idx]
		if len(child.children) == 0 && !child.isLeaf {
			// 删除该子节点
			parent.children = append(parent.children[:idx], parent.children[idx+1:]...)
		} else {
			break
		}
	}
}

// 查找
func (node *DomainNode[T]) lookup(domain string) (T, bool) {
	var empty T
	curr := node
	for part := range splitDomainReverseIterator(domain) {
		child, _ := curr.findChild(part)
		if child == nil {
			// 精确匹配失败，尝试通配符
			child, _ = curr.findChild("*")
			if child != nil && child.isLeaf {
				return child.data, true
			}
			return empty, false
		}
		curr = child
	}
	if curr.isLeaf {
		return curr.data, true
	}
	return empty, false
}

// each 遍历树中所有节点
func (node *DomainNode[T]) each(callback func(*DomainNode[T])) {
	callback(node)
	for _, child := range node.children {
		child.each(callback)
	}
}

package trie

import "slices"

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

// 二分查找，返回 (找到的节点, 下标)，如果没找到，下标为插入点
func (node *DomainNode[T]) findChild(label string) (*DomainNode[T], int) {
	low, high := 0, len(node.children)-1
	for low <= high {
		mid := (low + high) / 2
		if node.children[mid].label == label {
			return node.children[mid], mid
		} else if node.children[mid].label < label {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return nil, low // 未找到，low为插入点
}

// 插入时保持有序
func (node *DomainNode[T]) add(domain string, data T) {
	curr := node
	deep := node.deep
	for part := range splitDomainReverseIterator(domain) {
		child, idx := curr.findChild(part)
		if child == nil {
			child = newDomainNode[T](part, deep+1)
			// 插入到正确位置
			curr.children = append(curr.children, nil)
			copy(curr.children[idx+1:], curr.children[idx:])
			curr.children[idx] = child
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
		idx    int
	}
	var path []pathElem
	curr := node
	for part := range splitDomainReverseIterator(domain) {
		child, idx := curr.findChild(part)
		if child == nil {
			return
		}
		path = append(path, pathElem{curr, idx})
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
			parent.children = slices.Delete(parent.children, idx, idx+1)
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

package trie

type Node[T any] struct {
	parent   *Node[T]
	children map[string]*Node[T]
	data     T
	isLeaf   bool
}

func (n *Node[T]) IsLeaf() bool {
	return n.isLeaf
}

func (n *Node[T]) IsEmpty() bool {
	return len(n.children) == 0
}

func (n *Node[T]) Data() T {
	return n.data
}
func (n *Node[T]) Parent() *Node[T] {
	return n.parent
}

func (n *Node[T]) MarkAsLeaf() {
	n.isLeaf = true
}

func (n *Node[T]) MarkAsNode() {
	n.isLeaf = false
}

func (n *Node[T]) SetData(data T) {
	n.data = data
}

func (n *Node[T]) ForEach(f func(string, *Node[T])) {
	for k, v := range n.children {
		f(k, v)
	}
}

func (n *Node[T]) Has(k string) bool {
	_, ok := n.children[k]
	return ok
}

func (n *Node[T]) Get(k string) *Node[T] {
	return n.children[k]
}

func (n *Node[T]) GetWildcard() *Node[T] {
	if w, ok := n.children["*"]; ok {
		return w
	}
	return n.children[""]
}

func (n *Node[T]) Remove(k string) bool {
	_, ok := n.children[k]
	if ok {
		delete(n.children, k)
	}

	return ok
}

func (n *Node[T]) RemoveNode(node *Node[T]) {
	for k, v := range n.children {
		if v == node {
			delete(n.children, k)
			break
		}
	}
}

func (n *Node[T]) Set(k string, new *Node[T]) {
	n.children[k] = new
}

func (n *Node[T]) GetOrSet(k string, new *Node[T]) *Node[T] {
	node, ok := n.children[k]
	if ok {
		return node
	}
	n.children[k] = new
	return new
}

func NewNodeNil[T any](parent *Node[T], leaf ...bool) *Node[T] {
	return &Node[T]{
		parent:   parent,
		isLeaf:   len(leaf) > 0 && leaf[0],
		children: map[string]*Node[T]{},
	}
}

func NewNode[T any](data T, parent *Node[T], leaf ...bool) *Node[T] {
	return &Node[T]{
		parent:   parent,
		data:     data,
		isLeaf:   len(leaf) > 0 && leaf[0],
		children: map[string]*Node[T]{},
	}
}

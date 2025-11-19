package nodes

// BaseNode provides common functionality
type BaseNode struct {
	id  string
	typ string
}

func NewBaseNode(id, typ string) BaseNode {
	return BaseNode{id: id, typ: typ}
}

func (n *BaseNode) ID() string {
	return n.id
}

func (n *BaseNode) Type() string {
	return n.typ
}

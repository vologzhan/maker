package template

import "errors"

type FileType int

const (
	FileUnknown FileType = iota
	FileGo
)

type (
	Dir struct {
		Name          []Node
		Items         []Node
		Entry         bool
		parent        Node
		NextInKeyPath Fs
	}
	File struct {
		Type    FileType
		Name    []Node
		Content []Node
		Entry   bool
		parent  Node
	}
	Template struct {
		Items  []Node
		Entry  bool
		parent Node
	}
	Insert struct {
		Namespace string
		Name      string
		IsKey     bool
		ForMerge  bool
		Func      func(string) string
		Items     []Node
		parent    Node
	}
	Imports struct {
		Items  []*Import
		parent Node
	}
	Import struct {
		Name   []Node
		Alias  []Node
		Entry  bool
		parent Node
	}
	Word struct {
		Value  string
		parent Node
	}
	LineFeed struct {
		Value  int
		parent Node
	}
	Separator struct {
		Value  string
		parent Node
	}
	key struct {
		parent Node
	}
)

type Node interface {
	Parent() Node
}

func (n *Dir) Parent() Node       { return n.parent }
func (n *File) Parent() Node      { return n.parent }
func (n *Template) Parent() Node  { return n.parent }
func (n *Insert) Parent() Node    { return n.parent }
func (n *Imports) Parent() Node   { return n.parent }
func (n *Import) Parent() Node    { return n.parent }
func (n *Word) Parent() Node      { return n.parent }
func (n *LineFeed) Parent() Node  { return n.parent }
func (n *Separator) Parent() Node { return n.parent }
func (n *key) Parent() Node       { return n.parent }

type Fs interface {
	Node
	ParentFs() Fs
	NameNodes() []Node
}

func (n *File) ParentFs() Fs { return n.parent.(Fs) }
func (n *Dir) ParentFs() Fs {
	if n.parent == nil {
		return nil
	}
	return n.parent.(Fs)
}

func (n *File) NameNodes() []Node { return n.Name }
func (n *Dir) NameNodes() []Node  { return n.Name }

func UpToFsNode(n Node) (Fs, error) {
	for ; n != nil; n = n.Parent() {
		if fs, ok := n.(Fs); ok {
			return fs, nil
		}
	}

	return nil, errors.New("template: UpToFsNode: not found")
}

func IsEntry(tpl Node) bool {
	switch tpl := tpl.(type) {
	case *Dir:
		return tpl.Entry
	case *File:
		return tpl.Entry
	case *Template:
		return tpl.Entry
	case *Import:
		return tpl.Entry
	default:
		return false
	}
}

func IsChildOrCurrent(parent Node, child Node) bool {
	if parent == child {
		return true
	}

	for _, current := range getChildren(parent) {
		if IsChildOrCurrent(current, child) {
			return true
		}
	}

	return false
}

func getChildren(n Node) []Node {
	switch n := n.(type) {
	case *Dir:
		return append(n.Name, n.Items...)
	case *File:
		return append(n.Name, n.Content...)
	case *Template:
		return n.Items
	case *Insert:
		return n.Items
	case *Imports:
		out := make([]Node, len(n.Items))
		for i, item := range n.Items {
			out[i] = item
		}
		return out
	case *Import:
		return append(n.Name, n.Alias...)
	default:
		return nil
	}
}

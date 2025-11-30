package source

import (
	"fmt"
	"strings"
)

type Stringer interface {
	Node
	fmt.Stringer
}

func (n *Word) String() string      { return n.Value }
func (n *Separator) String() string { return n.Value }
func (n *LineFeed) String() string  { return strings.Repeat("\n", n.Value) }

func (n *Template) String() string {
	if n.Template.Entry {
		return fmt.Sprintf("%s\n", concat(n.Items))
	}
	if templateHasValue(n) {
		return concat(n.Items)
	}

	return ""
}

func templateHasValue(n Node) bool {
	for _, child := range GetChildren(n) {
		if childIns, ok := child.(*Insert); ok && childIns.Value != "" {
			return true
		}

		if templateHasValue(child) {
			return true
		}
	}

	return false
}

func (n *Insert) String() string {
	if n.Template.Func != nil {
		return n.Template.Func(n.Value)
	}

	if len(n.Template.Items) > 0 {
		if n.Value == "" {
			return ""
		}
		return concat(n.Items)
	}

	return n.Value
}

func (n *Imports) String() string {
	if len(n.Items) == 0 {
		return ""
	}
	if len(n.Items) == 1 {
		return fmt.Sprintf("import %s\n", n.Items[0].String())
	}

	buf := strings.Builder{}
	for _, item := range n.Items {
		buf.WriteString(fmt.Sprintf("\t%s\n", item.String()))
	}

	return fmt.Sprintf("import (\n%s)\n", buf.String())
}

func (n *Import) String() string {
	if len(n.Alias) > 0 {
		return fmt.Sprintf("%s \"%s\"", concat(n.Alias), concat(n.Name))
	}
	return fmt.Sprintf("\"%s\"", concat(n.Name))
}

func concat(nodes []Stringer) string {
	buf := strings.Builder{}
	for _, node := range nodes {
		buf.WriteString(node.String())
	}
	return buf.String()
}

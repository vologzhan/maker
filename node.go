package maker

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/vologzhan/maker/source"
	"github.com/vologzhan/maker/strcase"
	"github.com/vologzhan/maker/template"
)

type Node struct {
	id          uuid.UUID
	template    *template.Namespace
	parent      *Node
	children    map[string][]*Node
	entrypoints []source.Node
	inserts     map[string][]*source.Insert
	values      map[string]string
}

func New(tpl *template.Namespace, srcPath string) (*Node, error) {
	n := newNode(uuid.New(), tpl, nil, nil)

	tplEntry := tpl.Entrypoints[0].(*template.Dir)
	srcEntry := source.New(tplEntry, srcPath)

	n.entrypoints = append(n.entrypoints, srcEntry)
	n.values["path"] = srcPath

	return n, nil
}

func (n *Node) Id() uuid.UUID                  { return n.id }
func (n *Node) Parent() *Node                  { return n.parent }
func (n *Node) Values() map[string]string      { return n.values }
func (n *Node) ValueString(name string) string { return n.values[name] }
func (n *Node) ValueBool(name string) bool     { return n.values[name] != "" }

func (n *Node) SetValues(values map[string]string) error {
	if err := n.readPaths(); err != nil {
		return err
	}

	for name, value := range values {
		n.values[name] = value

		for _, ins := range n.inserts[name] {
			if err := ins.SetValue(value); err != nil {
				return err
			}
		}
	}

	return nil
}

func (n *Node) CreateChild(nspace string, id uuid.UUID, values map[string]string) (*Node, error) {
	tpl, ok := n.template.Children[nspace]
	if !ok {
		return nil, fmt.Errorf("maker: Node.CreateChild: child namespace '%s' does not exist", nspace)
	}

	if err := n.readKeys(tpl); err != nil {
		return nil, err
	}

	child := newNode(id, tpl, n, values)
	n.children[tpl.Name] = append(n.children[tpl.Name], child)

	for _, tplEntry := range tpl.Entrypoints {
		srcParent, err := findSourceByTemplate(n, tplEntry.Parent())
		if err != nil {
			return nil, err
		}

		srcEntry := source.CreateEntry(tplEntry, srcParent)
		child.entrypoints = append(child.entrypoints, srcEntry)

		if err := setInserts(child, srcEntry); err != nil {
			return nil, err
		}
	}

	return child, nil
}

func (n *Node) Children(nspace string) ([]*Node, error) {
	tpl, ok := n.template.Children[nspace]
	if !ok {
		return nil, fmt.Errorf("maker: Node.Children: child namespace '%s' does not exist", nspace)
	}

	if err := n.readKeys(tpl); err != nil {
		return nil, err
	}

	return n.children[nspace], nil
}

func (n *Node) Flush() error {
	for _, src := range n.entrypoints {
		if err := source.SaveRecursive(src); err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) getCurrentOrParent(nspace string) *Node {
	for ; n != nil; n = n.parent {
		if n.template.Name == nspace {
			return n
		}
	}
	return nil
}

func (n *Node) readKeys(tpl *template.Namespace) error {
	if _, ok := n.children[tpl.Name]; ok {
		return nil
	}

	n.children[tpl.Name] = []*Node{} // already read

	for _, tplKey := range tpl.Keys {
		if _, err := readSourceByTemplate(n, tplKey); err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) readPaths() error {
	for _, tpl := range n.template.Paths {
		node := n

		for {
			isFound, err := readSourceByTemplate(node, tpl)
			if err != nil {
				return err
			}
			if isFound {
				break
			}
			if node.parent == nil {
				break
			}

			node = node.parent
		}
	}

	return nil
}

func (n *Node) Delete() error {
	panic("implement me") // todo
}

func (n *Node) getEntryByTemplate(tpl template.Node) source.Node {
	for _, entry := range n.entrypoints {
		if tpl == entry.GetTemplate() {
			return entry
		}
	}
	return nil
}

func findSourceByTemplate(node *Node, tpl template.Node) (source.Node, error) {
	for _, parentTpl := range node.template.Entrypoints {
		if !template.IsChildOrCurrent(parentTpl, tpl) {
			continue
		}

		src := node.getEntryByTemplate(parentTpl)
		if src == nil {
			var err error
			src, err = findSourceByTemplate(node.parent, parentTpl)
			if err != nil {
				return nil, err
			}
			if src == nil {
				break
			}
		}

		foundNode := source.FindChildByTemplate(src, tpl)
		if foundNode != nil {
			return foundNode, nil
		}

		tplFs, err := template.UpToFsNode(tpl)
		if err != nil {
			return nil, err
		}

		isFound, err := readSourceByTemplate(node, tplFs)
		if err != nil {
			return nil, err
		}

		if !isFound {
			break
		}

		foundNode = source.FindChildByTemplate(src, tpl)
		if foundNode == nil {
			break
		}

		return foundNode, nil
	}

	return nil, errors.New("maker: findSourceByTemplate: node not found")
}

func readSourceByTemplate(node *Node, tpl template.Fs) (bool, error) {
	var isFound bool

	for _, entry := range node.entrypoints {
		if !template.IsChildOrCurrent(entry.GetTemplate(), tpl) {
			continue
		}

		sources, err := source.Read(entry.(source.Fs), tpl)
		if err != nil {
			return false, err
		}

		for _, src := range sources {
			if err := getInserts(node, src); err != nil {
				return false, err
			}
		}

		isFound = true
		break
	}

	return isFound, nil
}

func setInserts(node *Node, src source.Node) error {
	if ins, ok := src.(*source.Insert); ok {
		node = node.getCurrentOrParent(ins.Template.Namespace)
		if node == nil {
			return fmt.Errorf("maker: Node.setInserts: namespace '%s' does not exist", ins.Template.Namespace)
		}

		node.inserts[ins.Template.Name] = append(node.inserts[ins.Template.Name], ins)

		if err := ins.SetValue(node.values[ins.Template.Name]); err != nil {
			return err
		}
	}

	for _, child := range source.GetChildren(src) {
		if err := setInserts(node, child); err != nil {
			return err
		}
	}

	return nil
}

func getInserts(node *Node, src source.Node) error {
	if ins, ok := src.(*source.Insert); ok {
		nodeForInsert := node.getCurrentOrParent(ins.Template.Namespace)
		if nodeForInsert == nil {
			return fmt.Errorf("maker: Node.getInserts: namespace '%s' does not exist", ins.Template.Namespace)
		}
		nodeForInsert.inserts[ins.Template.Name] = append(nodeForInsert.inserts[ins.Template.Name], ins)
		if ins.Template.IsKey {
			nodeForInsert.values[ins.Template.Name] = ins.Value
		}
	}

	if template.IsEntry(src.GetTemplate()) {
		var ins *source.Insert
		for _, child := range source.GetChildren(src) {
			i, ok := child.(*source.Insert)
			if !ok || !i.Template.ForMerge {
				continue
			}
			if parent := node.getCurrentOrParent(i.Template.Namespace); parent != nil {
				continue
			}

			ins = i
			break
		}

		if ins == nil {
			return errors.New("engine: getInserts: namespace name not found")
		}

		var child *Node
		for _, ch := range node.children[ins.Template.Namespace] {
			if strcase.ToSnake(ch.values[ins.Template.Name]) == strcase.ToSnake(ins.Value) {
				child = ch
				break
			}
		}

		if child == nil {
			childTpl, ok := node.template.Children[ins.Template.Namespace]
			if !ok {
				return fmt.Errorf("engine: getInserts: child namespace '%s' does not exist", strcase.ToSnake(ins.Template.Namespace))
			}

			child = newNode(uuid.New(), childTpl, node, nil)
			node.children[ins.Template.Namespace] = append(node.children[ins.Template.Namespace], child)
		}

		child.entrypoints = append(child.entrypoints, src)
		node = child
	}

	for _, child := range source.GetChildren(src) {
		if err := getInserts(node, child); err != nil {
			return err
		}
	}

	return nil
}

func newNode(id uuid.UUID, tpl *template.Namespace, parent *Node, values map[string]string) *Node {
	if values == nil {
		values = make(map[string]string)
	}

	return &Node{
		id,
		tpl,
		parent,
		make(map[string][]*Node),
		nil,
		make(map[string][]*source.Insert),
		values,
	}
}

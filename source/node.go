package source

import (
	"errors"
	"fmt"
	slicesHelper "github.com/vologzhan/maker/helper/slices"
	"github.com/vologzhan/maker/template"
	"slices"
)

type (
	Dir struct {
		Template *template.Dir
		Parent   *Dir
		Name     []Stringer
		Items    []Fs
		RealName string
		Status   FsStatus
	}
	File struct {
		Template *template.File
		Parent   *Dir
		Name     []Stringer
		Content  []Stringer
		RealName string
		Status   FsStatus
	}
	Template struct {
		Template *template.Template
		Parent   Node
		Items    []Stringer
	}
	Insert struct {
		Template *template.Insert
		Parent   Node
		Value    string
		Items    []Stringer
	}
	Imports struct {
		Template *template.Imports
		Parent   *File
		Items    []*Import
	}
	Import struct {
		Template *template.Import
		Parent   *Imports
		Name     []Stringer
		Alias    []Stringer
	}
	Word struct {
		Template *template.Word
		Parent   Node
		Value    string
	}
	LineFeed struct {
		Template *template.LineFeed
		Parent   Node
		Value    int
	}
	Separator struct {
		Template *template.Separator
		Parent   Node
		Value    string
	}
)

type Node interface {
	GetTemplate() template.Node
	GetParent() Node
}

func (n *Dir) GetTemplate() template.Node {
	if n.Template == nil {
		return nil
	}
	return n.Template
}
func (n *File) GetTemplate() template.Node {
	if n.Template == nil {
		return nil
	}
	return n.Template
}
func (n *Template) GetTemplate() template.Node {
	if n.Template == nil {
		return nil
	}
	return n.Template
}
func (n *Insert) GetTemplate() template.Node {
	if n.Template == nil {
		return nil
	}
	return n.Template
}
func (n *Imports) GetTemplate() template.Node {
	if n.Template == nil {
		return nil
	}
	return n.Template
}
func (n *Import) GetTemplate() template.Node {
	if n.Template == nil {
		return nil
	}
	return n.Template
}
func (n *Word) GetTemplate() template.Node {
	if n.Template == nil {
		return nil
	}
	return n.Template
}
func (n *LineFeed) GetTemplate() template.Node {
	if n.Template == nil {
		return nil
	}
	return n.Template
}
func (n *Separator) GetTemplate() template.Node {
	if n.Template == nil {
		return nil
	}
	return n.Template
}

func (n *Dir) GetParent() Node       { return n.Parent }
func (n *File) GetParent() Node      { return n.Parent }
func (n *Template) GetParent() Node  { return n.Parent }
func (n *Insert) GetParent() Node    { return n.Parent }
func (n *Imports) GetParent() Node   { return n.Parent }
func (n *Import) GetParent() Node    { return n.Parent }
func (n *Word) GetParent() Node      { return n.Parent }
func (n *LineFeed) GetParent() Node  { return n.Parent }
func (n *Separator) GetParent() Node { return n.Parent }

func GetChildren(n Node) []Node {
	var children []Node
	switch n := n.(type) {
	case *Dir:
		for _, child := range n.Name {
			children = append(children, child)
		}
		for _, child := range n.Items {
			children = append(children, child)
		}
	case *File:
		for _, child := range n.Name {
			children = append(children, child)
		}
		for _, child := range n.Content {
			children = append(children, child)
		}
	case *Template:
		for _, child := range n.Items {
			children = append(children, child)
		}
	case *Insert:
		for _, child := range n.Items {
			children = append(children, child)
		}
	case *Imports:
		for _, child := range n.Items {
			children = append(children, child)
		}
	case *Import:
		for _, child := range n.Name {
			children = append(children, child)
		}
		for _, child := range n.Alias {
			children = append(children, child)
		}
	}

	return children
}

func FindChildByTemplate(src Node, tpl template.Node) Node {
	nearestNode, path := findNearestNodeAndPath(src, tpl)
	if len(path) > 0 {
		return nil
	}
	fs, ok := nearestNode.(Fs)
	if ok && fs.GetFsStatus() < FsStatusNotChanged {
		return nil
	}

	return nearestNode
}

func findNearestNodeAndPath(src Node, tpl template.Node) (Node, []template.Node) {
	var path []template.Node
	for ; tpl != nil; tpl = tpl.Parent() {
		if tpl != src.GetTemplate() {
			path = append(path, tpl)
			continue
		}

		slices.Reverse(path)

		for i, p := range path {
			var child Node

			switch src := src.(type) {
			case *Dir:
				for _, node := range src.Items {
					if node.GetTemplate() == p {
						child = node
						break
					}
				}
			case *File:
				for _, node := range src.Content {
					if node.GetTemplate() == p {
						child = node
						break
					}
				}
			default:
				panic(fmt.Sprintf("source: findChildByTemplate: unexpected parent node type '%T'", src))
			}

			if child == nil {
				return src, path[i:]
			}
			src = child
		}
	}

	return src, nil
}

func (n *Insert) SetValue(v string) error {
	n.Value = v

	fs, err := upToFsNode(n)
	if err != nil {
		return err
	}

	if fs.GetFsStatus() == FsStatusNotChanged {
		fs.SetFsStatus(FsStatusChanged)
	}

	if n.Template.Name != "type_go" {
		return nil
	}

	file, ok := fs.(*File)
	if !ok {
		return errors.New("source: Insert.SetValue: expected file node")
	}

	var imports *Imports
	for _, node := range file.Content {
		imports, ok = node.(*Imports)
		if ok {
			break
		}
	}

	if imports == nil {
		return errors.New("source: Insert.SetValue: not found node imports")
	}

	imports.AddImportByTypeGo(n.Value)

	return nil
}

func (n *Imports) AddImportByTypeGo(typeGo string) {
	var nameInImport string
	switch typeGo {
	case "uuid.UUID":
		nameInImport = "github.com/google/uuid"
	case "time.Time":
		nameInImport = "time"
	case "json.RawMessage":
		nameInImport = "encoding/json"
	default:
		return
	}

	for _, i := range n.Items {
		if nameInImport == concat(i.Name) {
			return
		}
	}

	imp := &Import{
		nil,
		n,
		make([]Stringer, 1),
		nil,
	}

	imp.Name[0] = &Word{
		nil,
		imp,
		nameInImport,
	}

	n.Items = append(n.Items, imp)
}

func DeleteNode(n Node) error {
	if fs, ok := n.(Fs); ok {
		switch fs.GetFsStatus() {
		case FsStatusNotRead, FsStatusNotExist:
			return nil
		case FsStatusNew:
			parentDir := fs.GetParentDir()
			slicesHelper.Delete(parentDir.Items, fs)
		default:
			fs.SetFsStatus(FsStatusDeleted)
		}
		return nil
	}

	switch parent := n.GetParent().(type) {
	case *Imports:
		slicesHelper.Delete(parent.Items, n.(*Import))
	case *Template:
		slicesHelper.Delete(parent.Items, n.(Stringer))
	case *File:
		slicesHelper.Delete(parent.Content, n.(Stringer))
	default:
		return fmt.Errorf("source: DeleteNode: unexpected node type '%T'", n)
	}

	fs, err := upToFsNode(n)
	if err != nil {
		return err
	}

	if fs.GetFsStatus() == FsStatusNotChanged {
		fs.SetFsStatus(FsStatusChanged)
	}

	return nil
}

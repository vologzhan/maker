package source

import (
	"fmt"
	"github.com/vologzhan/maker/template"
	"slices"
)

func New(tpl *template.Dir, path string) *Dir {
	dir := &Dir{tpl, nil, make([]Stringer, 1), nil, path, FsStatusNotRead}
	dir.Name[0] = &Insert{tpl.Name[0].(*template.Insert), dir, path, nil}

	return dir
}

func CreateEntry(tpl template.Node, parent Node) (Node, error) {
	entry := createNodeRecursive(tpl, parent, false)

	err := addEntryToParent(entry, parent)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func addEntryToParent(child, parent Node) error {
	switch parent := parent.(type) {
	case nil:
		// nothing
	case *Dir:
		parent.Items = append(parent.Items, child.(Fs))
	case *Imports:
		parent.Items = append(parent.Items, child.(*Import))
	case *File:
		tpl := child.GetTemplate()

		var tplPrev template.Node
		for _, n := range parent.Template.Content {
			if n == tpl {
				break
			}
			if _, ok := n.(*template.LineFeed); !ok {
				tplPrev = n
			}
		}

		for i := len(parent.Content) - 1; i >= 0; i-- {
			currentTpl := parent.Content[i].GetTemplate()
			if currentTpl == tpl {
				parent.Content = slices.Insert(parent.Content, i+1, child.(Stringer))
				return nil
			}
			if currentTpl == tplPrev {
				parent.Content = slices.Insert(parent.Content, i+2, child.(Stringer)) // вставка после перевода строки
				return nil
			}
		}

		return fmt.Errorf("source: addEntryToParent: in file no found place to insert node")
	default:
		return fmt.Errorf("source: addEntryToParent: unexpected parent node type '%T'", parent)
	}

	return nil
}

func createNodeRecursive(tpl template.Node, parent Node, skipEntry bool) Node {
	switch tpl := tpl.(type) {
	case *template.Dir:
		if tpl.Entry && skipEntry {
			return nil
		}
		parentDir, _ := parent.(*Dir)
		src := &Dir{
			tpl,
			parentDir,
			nil,
			nil,
			"",
			FsStatusNew,
		}
		for _, tpl := range tpl.Name {
			if newNode := createNodeRecursive(tpl, src, true); newNode != nil {
				src.Name = append(src.Name, newNode.(Stringer))
			}
		}
		for _, tpl := range tpl.Items {
			srcItem := createNodeRecursive(tpl, src, true)
			if srcItem == nil {
				continue
			}
			src.Items = append(src.Items, srcItem.(Fs))
		}
		return src
	case *template.File:
		if tpl.Entry && skipEntry {
			return nil
		}
		src := &File{
			tpl,
			parent.(*Dir),
			nil,
			nil,
			"",
			FsStatusNew,
		}
		for _, tpl := range tpl.Name {
			if newNode := createNodeRecursive(tpl, src, true); newNode != nil {
				src.Name = append(src.Name, newNode.(Stringer))
			}
		}
		for _, tpl := range tpl.Content {
			if newNode := createNodeRecursive(tpl, src, true); newNode != nil {
				src.Content = append(src.Content, newNode.(Stringer))
			}
		}
		return src
	case *template.Template:
		if tpl.Entry && skipEntry {
			return nil
		}
		src := &Template{tpl, parent, nil}
		for _, tpl := range tpl.Items {
			src.Items = append(src.Items, createNodeRecursive(tpl, src, true).(Stringer))
		}
		return src
	case *template.Insert:
		src := &Insert{tpl, parent, "", nil}
		if len(tpl.Items) > 0 {
			for _, tpl := range tpl.Items {
				src.Items = append(src.Items, createNodeRecursive(tpl, src, true).(Stringer))
			}
		}
		return src
	case *template.Imports:
		src := &Imports{Template: tpl, Parent: parent.(*File), Items: nil}
		for _, tpl := range tpl.Items {
			srcItem := createNodeRecursive(tpl, src, true)
			if srcItem == nil {
				continue
			}
			src.Items = append(src.Items, srcItem.(*Import))
		}
		return src
	case *template.Import:
		if tpl.Entry && skipEntry {
			return nil
		}
		src := &Import{tpl, parent.(*Imports), nil, nil}
		for _, tpl := range tpl.Name {
			src.Name = append(src.Name, createNodeRecursive(tpl, src, true).(Stringer))
		}
		for _, tpl := range tpl.Alias {
			src.Alias = append(src.Alias, createNodeRecursive(tpl, src, true).(Stringer))
		}
		return src
	case *template.Word:
		return &Word{tpl, parent, tpl.Value}
	case *template.LineFeed:
		return &LineFeed{tpl, parent, tpl.Value}
	case *template.Separator:
		return &Separator{tpl, parent, tpl.Value}
	default:
		panic(fmt.Sprintf("source: createNodeRecursive: unexpected node type '%T'", tpl))
	}
}

package source

import (
	"errors"
	"github.com/vologzhan/maker/template"
	"go/format"
	"golang.org/x/tools/imports"
	"os"
	"path"
	"slices"
)

type FsStatus int

const (
	FsStatusNotRead FsStatus = iota
	FsStatusNotExist
	FsStatusNotChanged
	FsStatusChanged
	FsStatusNew
)

type Fs interface {
	Node
	GetParentFs() Fs
	GetFsStatus() FsStatus
	SetFsStatus(status FsStatus)
	GetName() string
	GetRealName() string
	UpdateRealName()
}

func (n *Dir) GetParentFs() Fs {
	if n.Parent == nil {
		return nil
	}
	return n.Parent
}
func (n *File) GetParentFs() Fs        { return n.Parent }
func (n *Dir) GetFsStatus() FsStatus   { return n.Status }
func (n *File) GetFsStatus() FsStatus  { return n.Status }
func (n *Dir) SetFsStatus(s FsStatus)  { n.Status = s }
func (n *File) SetFsStatus(s FsStatus) { n.Status = s }
func (n *Dir) GetName() string         { return concat(n.Name) }
func (n *File) GetName() string        { return concat(n.Name) }
func (n *Dir) GetRealName() string     { return n.RealName }
func (n *File) GetRealName() string    { return n.RealName }
func (n *Dir) UpdateRealName()         { n.RealName = concat(n.Name) }
func (n *File) UpdateRealName()        { n.RealName = concat(n.Name) }

func SaveRecursive(node Node) error {
	fs, err := upToFsNode(node)
	if err != nil {
		return err
	}

	return saveRecursive(fs)
}

func saveRecursive(node Fs) error {
	var err error
	switch node := node.(type) {
	case *Dir:
		switch node.Status {
		case FsStatusNotRead, FsStatusNotExist, FsStatusNotChanged:
			// nothing
		case FsStatusChanged:
			err = rename(node)
		case FsStatusNew:
			err = createDir(node)
		}

		for _, child := range node.Items {
			if err := saveRecursive(child); err != nil {
				return err
			}
		}
	case *File:
		switch node.Status {
		case FsStatusNotRead, FsStatusNotExist, FsStatusNotChanged:
			// nothing
		case FsStatusChanged:
			err = updateFile(node)
		case FsStatusNew:
			err = createFile(node)
		}
	}

	if err != nil {
		return err
	}

	node.SetFsStatus(FsStatusNotChanged)

	return nil
}

func upToFsNode(node Node) (Fs, error) {
	for ; node != nil; node = node.GetParent() {
		if fs, ok := node.(Fs); ok {
			return fs, nil
		}
	}

	return nil, errors.New("source: upToFsNode: not found")
}

func buildPath(node Fs) string {
	var buf []string
	for ; node != nil; node = node.GetParentFs() {
		buf = append(buf, node.GetName())
	}
	slices.Reverse(buf)

	return path.Join(buf...)
}

func buildRealPath(node Fs) string {
	var buf []string
	for ; node != nil; node = node.GetParentFs() {
		buf = append(buf, node.GetRealName())
	}
	slices.Reverse(buf)

	return path.Join(buf...)
}

func rename(node Fs) error {
	if node.GetName() == node.GetRealName() {
		return nil
	}

	if err := os.Rename(buildRealPath(node), buildPath(node)); err != nil {
		return err
	}
	node.UpdateRealName()

	return nil
}

func createDir(node *Dir) error {
	err := os.Mkdir(buildPath(node), 0744)
	if err != nil {
		return err
	}

	node.UpdateRealName()

	return nil
}

func createFile(node *File) error {
	file, err := os.Create(buildPath(node))
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := buildContent(node)
	if err != nil {
		return err
	}

	_, err = file.Write(content)
	if err != nil {
		return err
	}

	node.UpdateRealName()

	return nil
}

func updateFile(f *File) error {
	if err := rename(f); err != nil {
		return err
	}

	realPath := buildRealPath(f)

	file, err := os.OpenFile(realPath, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := buildContent(f)
	if err != nil {
		return err
	}

	_, err = file.Write(content)
	if err != nil {
		return err
	}

	return nil
}

func buildContent(f *File) ([]byte, error) {
	content := []byte(concat(f.Content))

	switch f.Template.Type {
	case template.FileGo:
		formatedContent, err := format.Source(content)
		if err != nil {
			return nil, err
		}

		//imports.LocalPrefix = "" // todo

		return imports.Process("", formatedContent, nil)
	default:
		return content, nil
	}
}

package source

import (
	"errors"
	"fmt"
	slicesHelper "github.com/vologzhan/maker/helper/slices"
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
	FsStatusDeleted
)

type Fs interface {
	Node
	GetParentDir() *Dir
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
func (n *Dir) GetParentDir() *Dir {
	if n.Parent == nil {
		return nil
	}
	return n.Parent
}
func (n *File) GetParentDir() *Dir     { return n.Parent }
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
	switch n := node.(type) {
	case *Dir:
		if n.Status == FsStatusDeleted {
			return deleteDir(n)
		}
		if n.Status == FsStatusNew {
			if err := createDir(n); err != nil {
				return err
			}
		}
		if n.Status == FsStatusChanged {
			if err := rename(n); err != nil {
				return err
			}
		}

		for i := len(n.Items) - 1; i >= 0; i-- {
			if err := saveRecursive(n.Items[i]); err != nil {
				return err
			}
		}
	case *File:
		if n.Status == FsStatusDeleted {
			return deleteFile(n)
		}
		if n.Status == FsStatusNew {
			if err := createFile(n); err != nil {
				return err
			}
		}
		if n.Status == FsStatusChanged {
			if err := updateFile(n); err != nil {
				return err
			}
		}
	}

	if node.GetFsStatus() > FsStatusNotChanged {
		node.SetFsStatus(FsStatusNotChanged)
	}

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

func buildPath(node Fs) (string, error) {
	var buf []string
	for ; node != nil; node = node.GetParentFs() {
		name := node.GetName()
		if name == "" {
			return "", errors.New("source: buildPath: empty part of path")
		}

		buf = append(buf, name)
	}

	slices.Reverse(buf)

	return path.Join(buf...), nil
}

func buildRealPath(node Fs) (string, error) {
	var buf []string
	for ; node != nil; node = node.GetParentFs() {
		name := node.GetRealName()
		if name == "" {
			return "", errors.New("source: buildRealPath: empty part of path")
		}

		buf = append(buf, node.GetRealName())
	}

	slices.Reverse(buf)

	return path.Join(buf...), nil
}

func rename(node Fs) error {
	if node.GetName() == node.GetRealName() {
		return nil
	}

	oldPath, err := buildRealPath(node)
	if err != nil {
		return err
	}

	newPath, err := buildPath(node)
	if err != nil {
		return err
	}

	err = os.Rename(oldPath, newPath)
	if err != nil {
		return err
	}

	node.UpdateRealName()

	return nil
}

func createDir(node *Dir) error {
	newPath, err := buildPath(node)
	if err != nil {
		return err
	}

	err = os.Mkdir(newPath, 0744)
	if err != nil {
		return err
	}

	node.UpdateRealName()

	return nil
}

func createFile(node *File) error {
	newPath, err := buildPath(node)
	if err != nil {
		return err
	}

	file, err := os.Create(newPath)
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

func deleteDir(node *Dir) error {
	realPath, err := buildRealPath(node)
	if err != nil {
		return err
	}

	err = os.RemoveAll(realPath)
	if err != nil {
		return err
	}
	node.Parent.Items = slicesHelper.Delete(node.Parent.Items, Fs(node))

	return nil
}

func deleteFile(node *File) error {
	realPath, err := buildRealPath(node)
	if err != nil {
		return err
	}

	err = os.Remove(realPath)
	if err != nil {
		return err
	}
	node.Parent.Items = slicesHelper.Delete(node.Parent.Items, Fs(node))

	return nil
}

func updateFile(f *File) error {
	if err := rename(f); err != nil {
		return err
	}

	realPath, err := buildRealPath(f)
	if err != nil {
		return err
	}

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
			return nil, fmt.Errorf("source: buildContent: format source error: [%w], file: [%s], content:\n%s", err, f.GetName(), content)
		}

		//imports.LocalPrefix = "" // todo

		return imports.Process("", formatedContent, nil)
	default:
		return content, nil
	}
}

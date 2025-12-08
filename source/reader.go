package source

import (
	"fmt"
	"github.com/vologzhan/maker/template"
	"os"
	pathBase "path"
	"path/filepath"
	"strings"
)

var Test = false

func Read(src Fs, tpl template.Fs) ([]Node, error) {
	nearestNode, path := findNearestNodeAndPath(src, tpl)

	fsNode, ok := nearestNode.(Fs)
	if !ok {
		return nil, nil // уже прочитано, нода внутри файла
	}

	var pathFs []template.Fs
	for _, p := range path {
		if pFs, ok := p.(template.Fs); ok {
			pathFs = append(pathFs, pFs)
		} else {
			break
		}
	}

	return readPath(fsNode, pathFs)
}

func readPath(node Fs, path []template.Fs) ([]Node, error) {
	if node.GetFsStatus() == FsStatusNotExist {
		return nil, nil
	}

	var out []Node

	switch node := node.(type) {
	case *File:
		if node.Status != FsStatusNotRead {
			break
		}

		if err := readFile(node); err != nil {
			return nil, err
		}

		for _, item := range node.Content {
			out = append(out, item)
		}
	case *Dir:
		if node.Status == FsStatusNotRead {
			if err := readDir(node); err != nil {
				return nil, err
			}

			if len(out) == 0 {
				for _, item := range node.Items {
					out = append(out, item)
				}
			}
		}

		if len(path) == 0 {
			break
		}

		items := getDirItemsByTpl(node, path[0])

		for _, item := range items {
			if _, err := readPath(item, path[1:]); err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("source: readPath: unknown node type: %T", node)
	}

	return out, nil
}

func readFile(file *File) error {
	realPath, err := buildRealPath(file)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(realPath)
	if err != nil {
		return err
	}

	if err := parseContent(content, file); err != nil {
		return err
	}

	file.Status = FsStatusNotChanged

	return nil
}

func readDir(dir *Dir) error {
	realPath, err := buildRealPath(dir)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(realPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		var item Fs
		if entry.IsDir() {
			item = &Dir{nil, dir, nil, nil, entry.Name(), FsStatusNotRead}
		} else {
			item = &File{nil, dir, nil, nil, entry.Name(), FsStatusNotRead}
		}
		dir.Items = append(dir.Items, item)
	}

	tpls := getSortedDirTplItems(dir.Template.Items)

	for _, tpl := range tpls {
		items, err := getMatchedDirItemsWithoutTemplate(dir, tpl)
		if err != nil {
			return err
		}

		for _, src := range items {
			switch src := src.(type) {
			case *Dir:
				tpl := tpl.(*template.Dir)

				if tpl.NextInKeyPath != nil {
					hasNext, err := hasNextKeyPath(src, tpl)
					if err != nil {
						return err
					}
					if !hasNext {
						break
					}
				}

				src.Name = mustParse(src.RealName, tpl.NameNodes(), src)
				src.Template = tpl
			case *File:
				src.Name = mustParse(src.RealName, tpl.NameNodes(), src)
				src.Template = tpl.(*template.File)
			default:
				return fmt.Errorf("source: readDir: unexpected template node type: %T", tpl)
			}
		}

		if template.IsEntry(tpl) || len(items) > 0 {
			continue
		}

		var item Fs
		switch tpl := tpl.(type) {
		case *template.Dir:
			item = &Dir{tpl, dir, nil, nil, "", FsStatusNotExist}
		case *template.File:
			item = &File{tpl, dir, nil, nil, "", FsStatusNotExist}
		default:
			return fmt.Errorf("source: readDir: unexpected template node type: %T", tpl)
		}
		dir.Items = append(dir.Items, item)
	}

	dir.Status = FsStatusNotChanged

	return nil
}

func hasNextKeyPath(dir *Dir, tpl *template.Dir) (bool, error) {
	// todo могут быть ложные true, переделать на dry-run чтение
	var path []string

	nextTpl := tpl.NextInKeyPath
	for {
		path = append(path, buildSearchPattern(nextTpl))

		nextDir, ok := nextTpl.(*template.Dir)
		if !ok {
			break
		}
		if nextDir.NextInKeyPath == nil {
			break
		}

		nextTpl = nextDir.NextInKeyPath
	}

	realPath, err := buildRealPath(dir)
	if err != nil {
		return false, err
	}

	searchPattern := pathBase.Join(realPath, pathBase.Join(path...))

	matches, err := filepath.Glob(searchPattern)
	if err != nil {
		return false, err
	}

	return len(matches) > 0, nil
}

func getMatchedDirItemsWithoutTemplate(dir *Dir, tpl template.Fs) ([]Fs, error) {
	searchPattern := buildSearchPattern(tpl)
	_, tplIsDir := tpl.(*template.Dir)

	var out []Fs
	for _, item := range dir.Items {
		if item.GetTemplate() != nil {
			continue
		}

		_, itemIsDir := item.(*Dir)
		if itemIsDir != tplIsDir {
			continue
		}

		isMatch, err := pathBase.Match(searchPattern, item.GetRealName())
		if err != nil {
			return nil, err
		}
		if !isMatch {
			continue
		}

		out = append(out, item)

		if !template.IsEntry(tpl) {
			break
		}
	}

	return out, nil
}

func getSortedDirTplItems(nodes []template.Node) []template.Fs {
	fsNodes := make([]template.Fs, len(nodes))
	for i, node := range nodes {
		fsNodes[i] = node.(template.Fs)
	}

	var head []template.Fs
	var tail []template.Fs

	for _, node := range fsNodes {
		hasInsert := false
		for _, nameNode := range node.NameNodes() {
			if _, hasInsert = nameNode.(*template.Insert); hasInsert {
				break
			}
		}

		if hasInsert {
			tail = append(tail, node)
		} else {
			head = append(head, node)
		}
	}

	return append(head, tail...)
}

func getDirItemsByTpl(dir *Dir, tpl template.Fs) []Fs {
	var out []Fs
	for _, item := range dir.Items {
		if item.GetTemplate() == tpl {
			out = append(out, item)
		}
	}

	return out
}

func buildSearchPattern(tpl template.Fs) string {
	var buf strings.Builder

	for _, tpl := range tpl.NameNodes() {
		switch tpl := tpl.(type) {
		case *template.Insert:
			buf.WriteString("*")
		case *template.Word:
			buf.WriteString(tpl.Value)
		default:
			panic(fmt.Sprintf("source: buildSearchPattern: unexpected template node type '%T'", tpl))
		}
	}

	if _, ok := tpl.(*template.File); ok && Test {
		buf.WriteString(".e")
	}

	return buf.String()
}

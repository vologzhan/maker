package template

import "io/fs"

type Namespace struct {
	Name        string
	Parent      *Namespace
	Children    map[string]*Namespace
	Entrypoints []Node
	Paths       []Fs
	Keys        []Fs
	Values      map[string]bool
}

func New(fsys fs.FS, pathPrefix string) (*Namespace, error) {
	dir, err := parseDir(fsys, pathPrefix)
	if err != nil {
		return nil, err
	}

	dir.Name = []Node{
		&Insert{"", "path", true, true, nil, nil, nil},
	}
	dir.Entry = true

	nspace := &Namespace{
		"",
		nil,
		make(map[string]*Namespace),
		[]Node{dir},
		nil,
		nil,
		make(map[string]bool),
	}
	analyze(dir, nspace)

	return nspace, nil
}

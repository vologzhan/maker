package template

import "fmt"

func analyze(current Node, nspace *Namespace) {
	for _, child := range getChildren(current) {
		setParent(child, current)

		switch child := child.(type) {
		case *Insert:
			currentNspace := getCurrentOrParent(nspace, child.Namespace)
			if currentNspace == nil {
				nspace = getOrAddChild(nspace, child.Namespace)
				currentNspace = nspace
				addEntrypoint(nspace, current)
			}
			if child.IsKey {
				addKey(currentNspace, child)
			}
			addPath(currentNspace, child)
			currentNspace.Values[child.Name] = true
		case *key:
			addKeyForce(nspace, current.(Fs))
		default:
			analyze(child, nspace)
		}
	}
}

func setParent(n Node, parent Node) {
	switch n := n.(type) {
	case *Dir:
		n.parent = parent
	case *File:
		n.parent = parent
	case *Template:
		n.parent = parent
	case *Insert:
		n.parent = parent
	case *Imports:
		n.parent = parent
	case *Import:
		n.parent = parent
	case *Word:
		n.parent = parent
	case *LineFeed:
		n.parent = parent
	case *Separator:
		n.parent = parent
	case *key:
		n.parent = parent
	default:
		panic(fmt.Sprintf("template: setParent: unexpected node type '%T'", n))
	}
}

func getCurrentOrParent(nspace *Namespace, name string) *Namespace {
	for ; nspace != nil; nspace = nspace.Parent {
		if nspace.Name == name {
			break
		}
	}
	return nspace
}

func getOrAddChild(nspace *Namespace, name string) *Namespace {
	child, ok := nspace.Children[name]
	if !ok {
		child = &Namespace{
			name,
			nspace,
			make(map[string]*Namespace),
			nil,
			nil,
			nil,
			make(map[string]bool),
		}
		nspace.Children[name] = child
	}

	return child
}

func addEntrypoint(nspace *Namespace, node Node) {
	if file, ok := node.(*File); ok {
		nspace.Entrypoints = append(nspace.Entrypoints, file)
		file.Entry = true
		return
	}
	if dir, ok := node.(*Dir); ok {
		nspace.Entrypoints = append(nspace.Entrypoints, dir)
		dir.Entry = true
		return
	}

	for current := node; current != nil; current = current.Parent() {
		if IsEntry(current) {
			nspace.Entrypoints = append(nspace.Entrypoints, current)
			return
		}
	}
}

func addPath(nspace *Namespace, ins *Insert) {
	nspace.Paths = addToSliceOrReplaceLast(ins, nspace.Paths)
}

func addKey(nspace *Namespace, ins *Insert) {
	nspace.Keys = addToSliceOrReplaceLast(ins, nspace.Keys)
}

func addKeyForce(nspace *Namespace, fs Fs) {
	switch current := fs.(type) {
	case *Dir:
		current.Name = current.Name[1:]
	case *File:
		current.Name = current.Name[1:]
	default:
		panic(fmt.Sprintf("temlate: analyze: unexpected node type '%T'", current))
	}

	lastKey := nspace.Keys[len(nspace.Keys)-1]
	var path []Fs
	for current := fs; ; current = current.ParentFs() {
		path = append(path, current)

		if lastKey == current {
			break
		}
	}

	for i := len(path) - 1; i > 0; i-- {
		path[i].(*Dir).NextInKeyPath = path[i-1]
	}
}

func addToSliceOrReplaceLast(ins *Insert, slice []Fs) []Fs {
	nodeFs, err := UpToFsNode(ins)
	if err != nil {
		panic(err.Error())
	}

	if len(slice) == 0 {
		return append(slice, nodeFs)
	}

	lastItem := slice[len(slice)-1]
	for current := nodeFs.(Node); current != nil; current = current.Parent() {
		if lastItem == current {
			slice[len(slice)-1] = nodeFs
			return slice
		}
	}

	return append(slice, nodeFs)
}

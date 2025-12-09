package source

import (
	"fmt"
	"github.com/vologzhan/maker/template"
	"regexp"
	"strings"
)

const (
	stateBase = iota
	stateExpectedImport
	stateInImport
)

type templateLine struct {
	nodes    []template.Node
	lf       *template.LineFeed
	imports  *template.Imports
	tplEntry *template.Template
	next     *templateLine
}

type sourceLine struct {
	value string
	next  *sourceLine
}

func parseContent(content []byte, file *File) error {
	srcLine := splitSrcToLines(content)
	tplLine := splitTplToLines(file.Template.Content)
	var buf []string
	state := stateBase

	for ; srcLine != nil; srcLine = srcLine.next {
		if tplLine == nil {
			var tail []string
			for ; srcLine != nil; srcLine = srcLine.next {
				tail = append(tail, srcLine.value)
			}
			file.Content = append(file.Content, &Word{nil, file, strings.Join(tail, "\n")})

			break
		}

		if tplLine.imports != nil && state == stateBase {
			state = stateExpectedImport
		}

		switch state {
		case stateBase:
		case stateExpectedImport:
			if srcLine.value == "" {
				file.Content[len(file.Content)-1].(*LineFeed).Value++
				continue
			} else if !strings.HasPrefix(srcLine.value, "import") {
				err := parseImports(nil, tplLine.imports, tplLine.lf, file)
				if err != nil {
					return err
				}
				tplLine = tplLine.next
				state = stateBase
				break
			} else if strings.HasSuffix(srcLine.value, "(") {
				state = stateInImport
				continue
			} else {
				srcImp := strings.TrimPrefix(srcLine.value, "import")
				srcImp = strings.TrimSpace(srcImp)
				err := parseImports([]string{srcImp}, tplLine.imports, tplLine.lf, file)
				if err != nil {
					return err
				}
				tplLine = tplLine.next
				state = stateBase
				continue
			}
		case stateInImport:
			srcImp := strings.TrimSpace(srcLine.value)
			if srcImp == "" {
				continue
			}
			if srcImp != ")" {
				buf = append(buf, srcImp)
				continue
			}
			err := parseImports(buf, tplLine.imports, tplLine.lf, file)
			if err != nil {
				return err
			}
			buf = nil
			tplLine = tplLine.next
			state = stateBase
			continue
		}

		line, matched, err := parse(srcLine.value, tplLine.nodes, file)
		if err != nil {
			return err
		}

		if matched && tplLine.tplEntry != nil {
			file.Content = append(file.Content, &Template{tplLine.tplEntry, file, line})
			continue
		}
		if matched {
			file.Content = append(file.Content, append(line, &LineFeed{tplLine.lf, file, 1})...)
			tplLine = tplLine.next
			continue
		}
		if tplLine.tplEntry != nil && tplLine.next != nil {
			next := tplLine.next
			line, matched, err := parse(srcLine.value, next.nodes, file)
			if err != nil {
				return err
			}
			if matched {
				file.Content = append(file.Content, append(line, &LineFeed{next.lf, file, 1})...)
				tplLine = next.next
				continue
			}
		}
		if srcLine.value == "" && len(file.Content) == 0 {
			file.Content = append(file.Content, &LineFeed{nil, file, 1})
			continue
		}
		if srcLine.value == "" {
			if srcLf, ok := file.Content[len(file.Content)-1].(*LineFeed); ok {
				srcLf.Value++
				continue
			}
			file.Content = append(file.Content, &LineFeed{nil, file, 1})
		}
		file.Content = append(file.Content, append(line, &LineFeed{nil, file, 1})...)
	}

	return nil
}

func splitSrcToLines(content []byte) *sourceLine {
	lines := strings.Split(string(content), "\n")

	var first *sourceLine
	var prev *sourceLine

	for _, line := range lines {
		cur := &sourceLine{line, nil}
		if prev != nil {
			prev.next = cur
		}
		if first == nil {
			first = cur
		}
		prev = cur
	}

	return first
}

func splitTplToLines(content []template.Node) *templateLine {
	zeroLine := &templateLine{}
	prev := zeroLine
	var buf []template.Node
	var imps *template.Imports

	for _, node := range content {
		switch n := node.(type) {
		case *template.LineFeed:
			prev = newTemplateLine(buf, n, prev, imps, nil)
			buf = nil
			imps = nil
			continue
		case *template.Imports:
			imps = n
			continue
		case *template.Template:
			if n.Entry {
				prev = newTemplateLine(n.Items, nil, prev, nil, n)
				continue
			}
		}

		buf = append(buf, node)
	}

	if len(buf) > 0 || imps != nil {
		prev = newTemplateLine(buf, nil, prev, imps, nil)
	}

	return zeroLine.next
}

func newTemplateLine(nodes []template.Node, lf *template.LineFeed, prev *templateLine, imps *template.Imports, tplEntry *template.Template) *templateLine {
	tplLine := &templateLine{nodes, lf, imps, tplEntry, nil}
	if prev != nil {
		prev.next = tplLine
	}
	return tplLine
}

func parseImports(imps []string, tpl *template.Imports, tplLf *template.LineFeed, f *File) error {
	srcImps := &Imports{tpl, f, nil}
	f.Content = append(f.Content, srcImps)
	f.Content = append(f.Content, &LineFeed{tplLf, f, 1})

	for i := range imps {
		imps[i] = strings.Replace(imps[i], "\"", "", 2)
	}

	tplImps := getSortedTplImports(tpl)

	for _, imp := range imps {
		nameAndAlias := strings.Split(imp, " ")

		var name string
		var alias string
		if len(nameAndAlias) == 2 {
			name = nameAndAlias[1]
			alias = nameAndAlias[0]
		} else {
			name = nameAndAlias[0]
		}

		srcImp := &Import{nil, srcImps, nil, nil}
		srcImps.Items = append(srcImps.Items, srcImp)

		for i, tplImp := range tplImps {
			srcName, matched, err := parse(name, tplImp.Name, srcImp)
			if err != nil {
				return err
			}
			if !matched {
				continue
			}

			srcImp.Template = tplImp
			srcImp.Name = srcName

			if !tplImp.Entry {
				tplImps = append(tplImps[:i], tplImps[i+1:]...)
			}

			if len(alias) > 0 {
				srcAlias, _, err := parse(alias, tplImp.Alias, srcImp)
				if err != nil {
					return err
				}
				srcImp.Alias = srcAlias
			}

			break
		}

		if len(srcImp.Name) > 0 {
			continue
		}

		srcImp.Name = []Stringer{
			&Word{nil, srcImp, name},
		}

		if len(alias) == 0 {
			continue
		}

		srcImp.Alias = []Stringer{
			&Word{nil, srcImp, alias},
		}
	}

	return nil
}

func getSortedTplImports(imps *template.Imports) []*template.Import {
	var head []*template.Import
	var tail []*template.Import

	for _, imp := range imps.Items {
		hasInsert := false
		for _, nameNode := range imp.Name {
			_, hasInsert = nameNode.(*template.Insert)
			if hasInsert {
				break
			}
		}

		if hasInsert {
			tail = append(tail, imp)
		} else {
			head = append(head, imp)
		}
	}

	return append(head, tail...)
}

func mustParse(str string, tplNodes []template.Node, parent Node) []Stringer {
	nodes, matched, err := parse(str, tplNodes, parent)
	if err != nil {
		panic(fmt.Sprintf("source: mustParse: %s", err.Error()))
	}

	if !matched {
		panic("source: mustParse: no matches")
	}

	return nodes
}

func parse(str string, tplNodes []template.Node, parent Node) ([]Stringer, bool, error) {
	reg, err := compileRegexp(tplNodes)
	if err != nil {
		return nil, false, err
	}
	srcNodes, matched := matchRegexp(str, tplNodes, reg, parent)

	return srcNodes, matched, nil
}

func compileRegexp(tpls []template.Node) (*regexp.Regexp, error) {
	var buf strings.Builder
	buf.WriteString("^")
	for _, tplItem := range tpls {
		buf.WriteString(buildRegexp(tplItem))
	}
	if len(tpls) > 0 {
		if _, ok := tpls[len(tpls)-1].(*template.Word); ok {
			buf.WriteString("(.*)")
		}
	}
	buf.WriteString("$")

	return regexp.Compile(buf.String())
}

func buildRegexp(tpls template.Node) string {
	switch tpl := tpls.(type) {
	case *template.Word:
		return fmt.Sprintf("(%s)", regexp.QuoteMeta(tpl.Value))
	case *template.Separator:
		return `([\t ]+)`
	case *template.Insert:
		if len(tpl.Items) == 0 {
			return `([^/]+?)`
		}

		var buf strings.Builder
		for _, tplItem := range tpl.Items {
			buf.WriteString(buildRegexp(tplItem))
		}

		return fmt.Sprintf("(%s)?", buf.String())
	case *template.Template:
		var buf strings.Builder
		for _, tplItem := range tpl.Items {
			buf.WriteString(buildRegexp(tplItem))
		}
		return fmt.Sprintf("(%s)?", buf.String())
	default:
		panic(fmt.Sprintf("source: buildRegexp: unexpected template node type '%T'", tpl))
	}
}

func matchRegexp(str string, tplItems []template.Node, reg *regexp.Regexp, parent Node) ([]Stringer, bool) {
	match := reg.FindStringSubmatch(str)

	if len(match) < 1 {
		return []Stringer{
			&Word{
				nil,
				parent,
				str,
			},
		}, false
	}

	match = match[1:] // remove full match
	matchPos := 0

	var out []Stringer
	for _, tplItem := range tplItems {
		out = append(out, stringToNode(match, &matchPos, tplItem, parent))
	}

	if len(match) > matchPos && match[matchPos] != "" {
		out = append(out, &Word{
			nil,
			nil,
			match[matchPos],
		})
	}

	return out, true
}

func stringToNode(match []string, matchPos *int, tpl template.Node, parent Node) Stringer {
	switch tpl := tpl.(type) {
	case *template.Word:
		w := &Word{
			tpl,
			parent,
			tpl.Value,
		}
		*matchPos++
		return w
	case *template.Separator:
		sep := &Separator{
			tpl,
			parent,
			match[*matchPos],
		}
		*matchPos++
		return sep
	case *template.Insert:
		if len(tpl.Items) == 0 {
			ins := &Insert{
				tpl,
				parent,
				match[*matchPos],
				nil,
			}

			*matchPos++
			return ins
		}

		out := &Insert{
			tpl,
			parent,
			"",
			nil,
		}
		if match[*matchPos] != "" {
			out.Value = "1"
		}
		*matchPos++
		for _, tpl := range tpl.Items {
			out.Items = append(out.Items, stringToNode(match, matchPos, tpl, out))
		}
		return out
	case *template.Template:
		out := &Template{tpl, parent, nil}

		*matchPos++
		for _, tpl := range tpl.Items {
			out.Items = append(out.Items, stringToNode(match, matchPos, tpl, out))
		}
		return out
	default:
		panic(fmt.Sprintf("source: stringToNode: unexpected template node type '%T'", tpl))
	}
}

package source

import (
	"fmt"
	"github.com/vologzhan/maker/template"
	"regexp"
	"strings"
)

type strategyImports struct {
	template *template.Imports
	lineFeed *template.LineFeed
	file     *File
	rawLines []string
}

type strategyEntry struct {
	template *template.Template
	file     *File
	rawLines []string
}

func parseContent(content []byte, file *File) error {
	var buf []template.Node
	rawLines := strings.Split(string(content), "\n")

	var tplLines [][]template.Node
	for _, tpl := range file.Template.Content {
		buf = append(buf, tpl)

		if _, ok := tpl.(*template.LineFeed); ok {
			tplLines = append(tplLines, buf)
			buf = nil
			continue
		}

		if tplEntry, ok := tpl.(*template.Template); ok && tplEntry.Entry {
			tplLines = append(tplLines, buf)
			buf = nil
		}
	}

	var strategy interface{}
	tplPos := 0

	for i, rawLine := range rawLines {
		if tplPos >= len(tplLines) {
			file.Content = append(file.Content, &Word{nil, file, rawLine})
			if rawLine != "" || len(rawLines)-1 != i {
				file.Content = append(file.Content, &LineFeed{nil, file, 1})
			}
			continue // шаблон закончился, пропуск строк
		}

		if s, ok := strategy.(*strategyImports); ok {
			switch rawLine {
			case "":
			case ")":
				imps, err := parseImports(s)
				if err != nil {
					return err
				}

				file.Content = append(file.Content, imps...)
				strategy = nil
				tplPos++
			default:
				s.rawLines = append(s.rawLines, rawLine)
			}

			continue
		}

		if rawLine == "" {
			if len(rawLines)-1 == i {
				break
			}
			lf := file.Content[len(file.Content)-1].(*LineFeed)
			lf.Value++
			continue
		}

		tplLine, tplLf := prepareTemplateLine(tplLines[tplPos])

		switch tpl := tplLine[0].(type) {
		case *template.Template:
			if tpl.Entry {
				strategy = &strategyEntry{tpl, file, nil}

				tplPos++
				tplLine, tplLf = prepareTemplateLine(tplLines[tplPos])
			}
		case *template.Imports:
			strategyImps := &strategyImports{tpl, tplLf, file, nil}

			switch {
			case !strings.HasPrefix(rawLine, "import"):
				imps, err := parseImports(strategyImps)
				if err != nil {
					return err
				}
				file.Content = append(file.Content, imps...)

				tplPos++
				tplLine, tplLf = prepareTemplateLine(tplLines[tplPos])
			case strings.HasSuffix(rawLine, "\""):
				strategyImps.rawLines = append(strategyImps.rawLines, strings.TrimPrefix(rawLine, "import "))

				imps, err := parseImports(strategyImps)
				if err != nil {
					return err
				}
				file.Content = append(file.Content, imps...)

				tplPos++
				continue
			default:
				strategy = strategyImps
				continue
			}
		}

		srcLine, matched, err := parse(rawLine, tplLine, file)
		if err != nil {
			return err
		}

		if s, ok := strategy.(*strategyEntry); ok {
			if !matched {
				s.rawLines = append(s.rawLines, rawLine)
				continue
			}

			if err := parseEntry(s); err != nil {
				return err
			}
			strategy = nil
		}

		lf := &LineFeed{nil, file, 1}

		if matched {
			tplPos++
			lf.Template = tplLf
		}

		file.Content = append(file.Content, srcLine...)
		file.Content = append(file.Content, lf)
	}

	return nil
}

func prepareTemplateLine(tplLine []template.Node) ([]template.Node, *template.LineFeed) {
	lf, ok := tplLine[len(tplLine)-1].(*template.LineFeed)
	if ok {
		return tplLine[:len(tplLine)-1], lf
	}

	return tplLine, nil
}

func parseImports(strategy *strategyImports) ([]Stringer, error) {
	var strImps []string
	for _, rawLine := range strategy.rawLines {
		preparedLine := strings.TrimSpace(rawLine)
		preparedLine = strings.Replace(preparedLine, "\"", "", 2)
		strImps = append(strImps, preparedLine)
	}

	imps := &Imports{
		strategy.template,
		strategy.file,
		nil,
	}

	tplImps := getSortedTplImports(strategy.template)

	for _, strImp := range strImps {
		nameAndAlias := strings.Split(strImp, " ")

		var strName string
		var strAlias string
		if len(nameAndAlias) == 2 {
			strName = nameAndAlias[1]
			strAlias = nameAndAlias[0]
		} else {
			strName = nameAndAlias[0]
		}

		imp := &Import{
			nil,
			imps,
			nil,
			nil,
		}

		for i, tplImp := range tplImps {
			name, matched, err := parse(strName, tplImp.Name, imp)
			if err != nil {
				return nil, err
			}
			if !matched {
				continue
			}

			imp.Template = tplImp
			imp.Name = name

			if !tplImp.Entry {
				tplImps = append(tplImps[:i], tplImps[i+1:]...)
			}

			if len(strAlias) > 0 {
				imp.Alias = mustParse(strAlias, tplImp.Alias, imp)
			}

			break
		}

		if len(imp.Name) == 0 {
			imp.Name = []Stringer{
				&Word{nil, imp, strName},
			}

			if len(strAlias) > 0 {
				imp.Alias = []Stringer{
					&Word{nil, imp, strAlias},
				}
			}
		}

		imps.Items = append(imps.Items, imp)
	}

	lf := &LineFeed{
		Template: strategy.lineFeed,
		Parent:   strategy.file,
		Value:    1,
	}

	return []Stringer{imps, lf}, nil
}

func getSortedTplImports(imps *template.Imports) []*template.Import {
	var head []*template.Import
	var tail []*template.Import

	for _, imp := range imps.Items {
		hasInsert := false
		for _, nameNode := range imp.Name {
			if _, hasInsert = nameNode.(*template.Insert); hasInsert {
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

func parseEntry(strategy *strategyEntry) error {
	for _, rawLine := range strategy.rawLines {
		src, _, err := parse(rawLine, []template.Node{strategy.template}, strategy.file)
		if err != nil {
			return err
		}

		strategy.file.Content = append(strategy.file.Content, src...)
	}

	return nil
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
		out := &Template{
			Template: tpl,
			Parent:   parent,
			Items:    nil,
		}
		*matchPos++
		for _, tpl := range tpl.Items {
			out.Items = append(out.Items, stringToNode(match, matchPos, tpl, out))
		}
		return out
	default:
		panic(fmt.Sprintf("source: stringToNode: unexpected template node type '%T'", tpl))
	}
}

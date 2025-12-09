package template

import (
	"fmt"
	"github.com/vologzhan/maker/helper/strcase"
	"github.com/vologzhan/maker/template/lexer"
	"io/fs"
	"path/filepath"
	"strings"
)

func parseDir(fsys fs.FS, path string) (*Dir, error) {
	dir := &Dir{
		Name: parseName(filepath.Base(path)),
	}

	entries, err := fs.ReadDir(fsys, path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())

		var item Node
		if entry.IsDir() {
			item, err = parseDir(fsys, entryPath)
		} else {
			item, err = parseFile(fsys, entryPath)
		}

		dir.Items = append(dir.Items, item)
	}

	return dir, nil
}

func parseFile(fsys fs.FS, path string) (*File, error) {
	pathPrepared := strings.TrimSuffix(path, ".t")
	name := filepath.Base(pathPrepared)

	file := &File{
		Name: parseName(name),
	}

	content, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, err
	}

	switch filepath.Ext(name) {
	case ".go":
		file.Type = FileGo
		file.Content = parseContentGo(content)
	default:
		file.Type = FileUnknown
		file.Content = parseContent(content)
	}

	return file, nil
}

func parseName(name string) []Node {
	return parse(&strategyBase{
		lexer: lexer.NewPathLexer(name),
	})
}

func parseContent(content []byte) []Node {
	return parse(&strategyBase{
		lexer: lexer.NewContentLexer(string(content)),
	})
}

func parseContentGo(content []byte) []Node {
	return parse(&strategyFileGo{
		strategyBase: strategyBase{
			lexer: lexer.NewContentLexer(string(content)),
		},
	})
}

func parse(s strategy) []Node {
	var out interface{}
	for {
		tok := s.nextToken()
		out = s.handle(tok)
		if out != nil {
			break
		}
	}

	if out, ok := out.([]Node); ok {
		return out
	}

	panic(fmt.Sprintf("parse return value must be slice, actual: %T", out))
}

type strategy interface {
	handle(token lexer.Token) interface{}
	nextToken() lexer.Token
}

type strategyBase struct {
	buf   []Node
	child strategy
	lexer *lexer.Lexer
}

func (s *strategyBase) nextToken() lexer.Token { return s.lexer.NextToken() }

func (s *strategyBase) handle(tok lexer.Token) interface{} {
	if s.child != nil {
		item := s.child.handle(tok)
		if item != nil {
			s.buf = append(s.buf, item.(Node))
			s.child = nil
		}

		return nil
	}

	switch token := tok.(type) {
	case nil:
		return s.buf
	case lexer.Word:
		s.buf = append(s.buf, &Word{string(token), nil})
	case lexer.LineFeed:
		s.buf = append(s.buf, &LineFeed{int(token), nil})
	case lexer.Separator:
		s.buf = append(s.buf, &Separator{string(token), nil})
	case lexer.TemplateStart:
		s.child = &strategyTemplate{
			strategyBase: strategyBase{
				lexer: s.lexer,
			},
		}
	case lexer.TemplateChildStart:
		s.child = &strategyTemplateEntry{
			strategyBase: strategyBase{
				lexer: s.lexer,
			},
		}
	default:
		panic(fmt.Sprintf("strategy base, unexpected token type: '%T'", token))
	}

	return nil
}

type strategyTemplate struct {
	strategyBase
	state  stateTemplate
	hasKey bool
}

type stateTemplate int

const (
	stateInsertWithFunc stateTemplate = iota
	stateInsert
	stateInsertWithCondition
	stateSimpleTemplate
)

func (s *strategyTemplate) handle(tok lexer.Token) interface{} {
	if s.child != nil {
		item := s.child.handle(tok)
		if item != nil {
			s.buf = append(s.buf, item.(Node))
			s.child = nil
		}

		if _, ok := item.(*Insert); ok {
			s.state = stateSimpleTemplate
		}

		return nil
	}

	switch token := tok.(type) {
	case lexer.Word, lexer.Separator, lexer.LineFeed, lexer.TemplateStart:
		s.strategyBase.handle(token)
	case lexer.TemplateKey:
		s.hasKey = true
	case lexer.TemplateSeparator:
		s.state = stateInsert
	case lexer.TemplateCondition:
		s.state = stateInsertWithCondition
	case lexer.TemplateEnd:
		switch s.state {
		case stateInsert:
			name := strcase.ToSnake(s.buf[1].(*Word).Value)
			return &Insert{
				strcase.ToSnake(s.buf[0].(*Word).Value),
				name,
				s.hasKey,
				isInsertForMerge(name),
				nil,
				nil,
				nil,
			}
		case stateInsertWithCondition:
			name := strcase.ToSnake(s.buf[1].(*Word).Value)
			return &Insert{
				strcase.ToSnake(s.buf[0].(*Word).Value),
				name,
				s.hasKey,
				isInsertForMerge(name),
				nil,
				s.buf[2:],
				nil,
			}
		case stateSimpleTemplate:
			return &Template{s.buf, false, nil}
		case stateInsertWithFunc:
			if s.hasKey && len(s.buf) == 0 {
				return &key{}
			}

			insertBuf := strings.Builder{}
			for _, item := range s.buf {
				if w, ok := item.(*Word); ok {
					insertBuf.WriteString(w.Value)
				} else {
					panic("insert, unexpected token type")
				}
			}

			nspace, name, f := splitToNamespaceNameFunction(insertBuf.String())

			return &Insert{
				nspace,
				name,
				s.hasKey,
				isInsertForMerge(name),
				f,
				nil,
				nil,
			}
		}
	default:
		panic(fmt.Sprintf("state template, unexpected token type: '%T'", token))
	}

	return nil
}

type strategyTemplateEntry struct {
	strategyBase
}

func (s *strategyTemplateEntry) handle(tok lexer.Token) interface{} {
	if s.child != nil {
		item := s.child.handle(tok)
		if item != nil {
			s.buf = append(s.buf, item.(Node))
			s.child = nil
		}

		return nil
	}

	switch token := tok.(type) {
	case lexer.Word, lexer.Separator, lexer.LineFeed, lexer.TemplateStart:
		s.strategyBase.handle(token)
	case lexer.TemplateChildEnd:
		lf, ok := s.lexer.NextToken().(lexer.LineFeed)
		if !ok {
			panic("entry template must end with line feed")
		}
		s.lexer.GoBack(int(lf) - 1)

		return &Template{s.buf, true, nil}
	default:
		panic(fmt.Sprintf("strategy template entry, unexpected token type: '%T'", token))
	}

	return nil
}

type strategyFileGo struct {
	strategyBase
	state stateFileGo
}

type stateFileGo int

const (
	statePackage stateFileGo = iota
	stateExpectedWordImport
	stateNormal
)

func (s *strategyFileGo) handle(tok lexer.Token) interface{} {
	if s.child != nil {
		item := s.child.handle(tok)
		if item != nil {
			s.buf = append(s.buf, item.(Node))

			if _, ok := item.(*Imports); ok {
				s.state = stateNormal
			}

			s.child = nil
		}

		return nil
	}

	switch token := tok.(type) {
	case lexer.Separator, lexer.TemplateStart, lexer.TemplateChildStart:
		return s.strategyBase.handle(tok)
	case lexer.Word:
		if s.state == stateExpectedWordImport {
			if token == "import" {
				s.child = &strategyImport{
					strategyBase: strategyBase{
						lexer: s.lexer,
					},
					imps: &Imports{},
				}
				break
			}

			s.buf = append(s.buf, &Imports{})
			s.buf = append(s.buf, &LineFeed{1, nil})
			s.state = stateNormal
		}
		s.strategyBase.handle(tok)
	case lexer.LineFeed:
		if s.state == statePackage {
			s.state = stateExpectedWordImport
		}

		s.strategyBase.handle(tok)
	case nil:
		if s.state == stateExpectedWordImport {
			s.buf = append(s.buf, &Imports{})
			s.buf = append(s.buf, &LineFeed{1, nil})
		}

		return s.strategyBase.handle(tok)
	default:
		panic(fmt.Sprintf("state file go, unexpected token type: '%T'", token))
	}

	return nil
}

type strategyImport struct {
	strategyBase
	state stateImport
	imps  *Imports
	alias []Node
}

type stateImport int

const (
	stateSingleLineImport stateImport = iota
	stateMultilineImport
)

func (s *strategyImport) handle(tok lexer.Token) interface{} {
	if s.child != nil {
		item := s.child.handle(tok)
		if item != nil {
			if tpl, ok := item.(*Template); ok && tpl.Entry {
				s.appendEntry(tpl)
			} else {
				s.buf = append(s.buf, item.(Node))
			}

			s.child = nil
		}

		return nil
	}

	switch token := tok.(type) {
	case lexer.Word:
		if s.state == stateSingleLineImport && token == "(" {
			s.state = stateMultilineImport
			break
		}
		if s.state == stateMultilineImport && token == ")" {
			return s.imps
		}

		tokenStr := string(token)
		if strings.HasPrefix(tokenStr, `"`) {
			tokenStr = strings.TrimPrefix(tokenStr, `"`)
		}
		if strings.HasSuffix(tokenStr, `"`) {
			tokenStr = strings.TrimSuffix(tokenStr, `"`)
		}
		token = lexer.Word(tokenStr)

		s.strategyBase.handle(token)
	case lexer.Separator:
		if len(s.buf) > 0 {
			s.alias = s.buf
			s.buf = nil
		}
	case lexer.LineFeed:
		if s.state == stateSingleLineImport {
			return &Imports{
				[]*Import{
					{
						s.buf,
						s.alias,
						false,
						nil,
					},
				},
				nil,
			}
		}
		if len(s.buf) > 0 {
			s.imps.Items = append(s.imps.Items, &Import{
				Name:  s.buf,
				Alias: s.alias,
			})
			s.buf = nil
			s.alias = nil
		}
	case lexer.TemplateStart, lexer.TemplateChildStart:
		s.strategyBase.handle(token)
	default:
		panic(fmt.Sprintf("state import, unexpected token type: '%T'", token))
	}

	return nil
}

func (s *strategyImport) appendEntry(tpl *Template) {
	var name []Node
	var alias []Node

	for _, item := range tpl.Items {
		switch item := item.(type) {
		case *Separator:
			if len(name) > 0 {
				alias = name
				name = nil
			}
			continue
		case *Word:
			str := item.Value
			if strings.HasPrefix(str, `"`) {
				str = strings.TrimPrefix(str, `"`)
			}
			if strings.HasSuffix(str, `"`) {
				str = strings.TrimSuffix(str, `"`)
			}
			if len(str) > 0 {
				name = append(name, &Word{str, nil})
			}
		default:
			name = append(name, item)
		}
	}
	s.imps.Items = append(s.imps.Items, &Import{name, alias, true, nil})
}

func isInsertForMerge(name string) bool {
	return name == "name"
}

func splitToNamespaceNameFunction(s string) (string, string, func(string) string) {
	parts := strings.Split(strcase.ToSnake(s), "_")

	namespace := parts[0]
	name := strings.Join(parts[1:], "_")

	caseMap := map[string]func(string) string{
		strcase.ToCamel(s):         strcase.ToCamel,
		strcase.ToKebab(s):         strcase.ToKebab,
		strcase.ToPascal(s):        strcase.ToPascal,
		strcase.ToSnake(s):         strcase.ToSnake,
		strcase.ToScreamerSnake(s): strcase.ToScreamerSnake,
		strcase.ToSentence(s):      strcase.ToSentence,
	}

	f, ok := caseMap[s]
	if !ok {
		panic(fmt.Sprintf("unexpected case of '%s'", s))
	}

	return namespace, name, f
}

package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

func NewPathLexer(buf string) *Lexer {
	return &Lexer{
		buf: []rune(buf),
		pos: 0,
		words: map[rune]Token{
			'{': TemplateStart("{"),
			'}': TemplateEnd("}"),
			'!': TemplateKey("!"),
		},
	}
}

func NewContentLexer(buf string) *Lexer {
	return &Lexer{
		buf: []rune(buf),
		pos: 0,
		words: map[rune]Token{
			'▶': TemplateStart("▶"),
			'◀': TemplateEnd("◀"),
			'⏩': TemplateChildStart("⏩"),
			'⏪': TemplateChildEnd("⏪"),
			'↔': TemplateCondition("↔"),
			'➡': TemplateSeparator("➡"),
			'⬇': TemplateKey("⬇"),
		},
	}
}

type Lexer struct {
	buf   []rune
	pos   int
	words map[rune]Token
}

type (
	Word               string
	Separator          string
	LineFeed           int
	TemplateStart      string
	TemplateEnd        string
	TemplateChildStart string
	TemplateChildEnd   string
	TemplateKey        string
	TemplateSeparator  string
	TemplateCondition  string
)

type Token interface {
	fmt.Stringer
}

func (w Word) String() string               { return string(w) }
func (s Separator) String() string          { return string(s) }
func (lf LineFeed) String() string          { return strings.Repeat("\n", int(lf)) }
func (t TemplateStart) String() string      { return string(t) }
func (t TemplateEnd) String() string        { return string(t) }
func (t TemplateChildStart) String() string { return string(t) }
func (t TemplateChildEnd) String() string   { return string(t) }
func (t TemplateKey) String() string        { return string(t) }
func (t TemplateSeparator) String() string  { return string(t) }
func (t TemplateCondition) String() string  { return string(t) }

const (
	stateStart = iota
	stateEnd
	stateWord
	stateTemplateSymbol
	stateLineFeed
	stateSeparator
)

func (l *Lexer) NextToken() Token {
	state := stateStart
	var buf []rune

	for {
		current, ok := l.nextRune()

		var newState int

		switch {
		case !ok:
			newState = stateEnd
		case current == ' ' || current == '\t':
			newState = stateSeparator
		case current == '\n':
			newState = stateLineFeed
		case unicode.IsLetter(current) || unicode.IsDigit(current):
			newState = stateWord
		default:
			if _, ok := l.words[current]; ok {
				newState = stateTemplateSymbol
			} else {
				newState = stateWord
			}
		}

		if state == stateStart {
			state = newState
		}

		if state != newState && newState != stateEnd {
			l.GoBack(1)
		}

		switch state {
		case stateWord:
			if state != newState {
				return Word(buf)
			}
			buf = append(buf, current)
		case stateTemplateSymbol:
			w, _ := l.words[current]
			return w
		case stateSeparator:
			if state != newState {
				return Separator(buf)
			}
			buf = append(buf, current)
		case stateLineFeed:
			if state != newState {
				return LineFeed(len(buf))
			}
			buf = append(buf, current)
		case stateEnd:
			return nil
		case stateStart:
			panic("unreachable state 'start'")
		}
	}
}

func (l *Lexer) GoBack(i int) {
	l.pos -= i
}

func (l *Lexer) nextRune() (rune, bool) {
	if l.pos >= len(l.buf) {
		return ' ', false
	}

	out := l.buf[l.pos]
	l.pos++

	return out, true
}

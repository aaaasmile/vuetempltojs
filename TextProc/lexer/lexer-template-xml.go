package lexer

import (
	"fmt"
	"io/ioutil"
	"strings"
	"unicode"
	"unicode/utf8"
)

type itemType int

type VueTempl struct {
	TokenTag string
}

const (
	eof                = -1
	itemError itemType = iota
	itemEOF
	itemSubTagContent
	itemTagContent
	itemText
)

type item struct {
	typ itemType

	val string
}

type lexer struct {
	name     string
	input    string
	tokentag string
	start    int
	pos      int
	width    int
	state    stateFn
	items    chan item
}

type stateFn func(*lexer) stateFn

func (l *lexer) run() {
	for state := lexText; state != nil; {
		state(l)
	}
	close(l.items)
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}

	ru, s := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = s
	l.pos += l.width
	return ru
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{
		itemError,
		fmt.Sprintf(format, args...),
	}
	return nil
}

func (l *lexer) peek() rune {
	ru := l.next()
	l.backup()
	return ru
}

func (l *lexer) nextItem() item {
	for {
		select {
		case item := <-l.items:
			return item
		default:
			l.state = l.state(l)
		}
	}
}

func lexDocInTag(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof || r == '\n':
			return l.errorf("string error")
		case r == '>':
			l.backup()
			l.emit(itemSubTagContent)
			return nil
		}
	}
}

func lexInsideTagContent(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof || r == '\n':
			return l.errorf("format error")
		case unicode.IsSpace(r):
			l.ignore()
		case r == '/':
			l.ignore()
			return lexDocInTag
		}
	}
}

func lexTagContent(l *lexer) stateFn {
	l.pos += len(l.tokentag)
	l.emit(itemTagContent)
	return lexInsideTagContent
}

func lexText(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], l.tokentag) {
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexTagContent // Next state
		}
		if l.next() == eof {
			break
		}
	}
	if l.pos > l.start {
		l.emit(itemText)
	}
	return nil
}

func lexCtor(name, input string, tt string) *lexer {
	l := &lexer{
		name:     name,
		input:    input,
		tokentag: tt,
		state:    lexText,
		items:    make(chan item, 2),
	}
	return l
}

func (vt *VueTempl) GetTemplateContent(str string) string {
	if vt.TokenTag != "" {
		vt.TokenTag = "<template>"
	}

	l := lexCtor("Text lex", str, vt.TokenTag)
	rr := ""
	for {
		item := l.nextItem()
		fmt.Printf("type %v, val %q\n", item.typ, item.val)
		if item.typ == itemSubTagContent {
			rr = item.val
		}
		if l.state == nil {
			break
		}
	}
	return rr

}

func (vt *VueTempl) GetTemplateFromFile(filename string) (string, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	s := string(buf)
	//fmt.Println(s)
	vn := vt.GetTemplateContent(s)
	if vn == "" {
		return "", fmt.Errorf("Template is empty")
	}
	return vn, nil
}

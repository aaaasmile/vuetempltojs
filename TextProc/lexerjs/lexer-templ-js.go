package lexerjs

import (
	"fmt"
	"io/ioutil"
	"strings"
	"unicode"
	"unicode/utf8"
)

type itemType int

const (
	eof               = -1
	itemText itemType = iota
	itemTemplName
	itemTagName
	itemTemplNameChildContent
	itemTemplAfter
	itemError
	itemEOF
)

type item struct {
	typ itemType

	val string
}

type lexer struct {
	name       string
	input      string
	sectionkey string
	start      int
	pos        int
	width      int
	state      stateFn
	items      chan item
}

type stateFn func(*lexer) stateFn

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

func lexOutsideTemplString(l *lexer) stateFn {
	for {
		if l.next() == eof {
			l.backup()
			l.emit(itemTemplAfter)
			return nil
		}
	}
}

func lexInsideTemplString(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			return l.errorf("Template string is not terminated")
		case r == '`':
			l.ignore()
			return lexOutsideTemplString
		}
	}
}

func lexInsideTemplate(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			return l.errorf("Template section format error")
		case unicode.IsSpace(r):
			l.ignore()
		case r == '`':
			l.ignore()
			if l.pos > l.start {
				l.emit(itemTemplNameChildContent)
			}
			return lexInsideTemplString
		}
	}
}

func lexStateTagName(l *lexer) stateFn {
	l.pos += len(l.sectionkey)
	l.emit(itemTagName)
	return lexInsideTemplate
}

func lexStateText(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], l.sectionkey) {
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexStateTagName
		}
		if l.next() == eof {
			return l.errorf("Template section not found")
		}
	}
	if l.pos > l.start {
		l.emit(itemText)
	}
	return nil
}

func lexCtor(name, input string, tt string) *lexer {
	l := &lexer{
		name:       name,
		input:      input,
		sectionkey: tt,
		state:      lexStateText,
		items:      make(chan item, 2),
	}
	return l
}

type JsTempl struct {
	SectionKey     string
	BeforeText     string
	SectionText    string
	SectionContent string
	AfterText      string
}

func (vt *JsTempl) String() string {
	return fmt.Sprintf("%s%s `%s`%s", vt.BeforeText, vt.SectionKey, vt.SectionContent, vt.AfterText)
}

func (vt *JsTempl) ParseComponentContent(str string) error {
	if vt.SectionKey == "" {
		vt.SectionKey = "template:"
	}

	l := lexCtor("Text lex", str, vt.SectionKey)
	for {
		item := l.nextItem()
		//fmt.Printf("*** type %v, val %q\n", item.typ, item.val)
		switch item.typ {
		case itemError:
			return fmt.Errorf(item.val)
		case itemText:
			vt.BeforeText = item.val
		case itemTemplAfter:
			vt.AfterText = item.val
		case itemTagName:
			vt.SectionText = item.val
		}
		if l.state == nil {
			break
		}
	}
	return nil

}

func (vt *JsTempl) SplitContentComponent(filename string) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	s := string(buf)
	//fmt.Println(s)
	return vt.ParseComponentContent(s)
}

func (vt *JsTempl) WriteFile(filename string) error {
	buf := []byte(vt.String())
	return ioutil.WriteFile(filename, buf, 0666)
}

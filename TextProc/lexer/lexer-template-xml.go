package lexer

import (
	"fmt"
	"io/ioutil"
	"strings"
	"unicode/utf8"
)

type itemType int

const (
	eof               = -1
	itemText itemType = iota
	itemTagName
	itemTagChildContent
	itemError
	itemEOF
)

type item struct {
	typ itemType

	val string
}

type lexer struct {
	name        string
	input       string
	tokentag    string
	endtokentag string
	start       int
	pos         int
	width       int
	state       stateFn
	items       chan item
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

func lexInsideTagContent(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], l.endtokentag) {
			//fmt.Println("** end of tag", l.input[l.pos:])
			if l.pos > l.start {
				l.backup() // in questo punto si è posizionati sul primo carattere di endtokentag
				l.emit(itemTagChildContent)
				return nil
			}
			return l.errorf("Lex is wrong on tag %s", l.endtokentag)
		}
		if l.next() == eof {
			return l.errorf("Malformed file, end of tag %s not found", l.endtokentag)
		}
	}
}

func lexTagContent(l *lexer) stateFn {
	l.pos += len(l.tokentag)
	l.emit(itemTagName)
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

func lexCtor(name, input string, tt string, endtt string) *lexer {
	l := &lexer{
		name:        name,
		input:       input,
		tokentag:    tt,
		endtokentag: endtt,
		state:       lexText,
		items:       make(chan item, 2),
	}
	return l
}

type VueTempl struct {
	TokenTag    string
	EndTokenTag string
}

func (vt *VueTempl) GetTemplateContent(str string) (string, error) {
	if vt.TokenTag == "" {
		vt.TokenTag = "<template>"
		vt.EndTokenTag = "</template>"
	}
	if vt.EndTokenTag == "" {
		return "", fmt.Errorf("Lex not properly configured")
	}

	l := lexCtor("Text lex", str, vt.TokenTag, vt.EndTokenTag)
	rr := ""
	for {
		item := l.nextItem()
		fmt.Printf("type %v, val %q\n", item.typ, item.val)
		if item.typ == itemError {
			return "", fmt.Errorf(item.val)
		}
		if item.typ == itemTagChildContent {
			rr = item.val
		}
		if l.state == nil {
			break
		}
	}
	return rr, nil

}

func (vt *VueTempl) GetTemplateFromFile(filename string) (string, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	s := string(buf)
	//fmt.Println(s)
	return vt.GetTemplateContent(s)
}

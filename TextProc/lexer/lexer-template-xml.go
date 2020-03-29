package lexer

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"unicode/utf8"
)

// Il lexer ha degli stati dove vengono emessi degli eventi. Gli eventi sono del tipo itemType ed hanno un valore val.
// Il principio è molto semplice. Viene scansionata la stringa di input carattere per carattere partendo dallo stato
// lexStateText. Lo stato è una funzione e ritorna lo stato successivo (una funzione del tipo stateFn). Se non ce ne sono di altri, allora nil.
// Quando il lexer in una funzione di stato riconosce qualcosa di
// interessante, allora chiama la funzione emit. Essa passa nel channel di comunicazione, un item che è la stringa di testo
// scansionato tra la posizione corrente (variabile pos) e start.
// Il chiamante del lexer, che è la funzione GetTemplateContent(), esegue in un ciclo la chiamata a nextItem() fino a quando lo stato è nil.

type itemType int

const (
	eof               = -1
	itemText itemType = iota
	itemTagName
	itemTagChildContent
	itemError
	itemEOF
	itemTextWrong
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

func lexStateChild(l *lexer) stateFn {
	endTagCount := 1
	templToken := strings.TrimRight(l.tokentag, ">")
	for {
		if strings.HasPrefix(l.input[l.pos:], templToken) {
			//fmt.Println("*** Sub token found on level ", endTagCount)
			endTagCount++
		}
		if strings.HasPrefix(l.input[l.pos:], l.endtokentag) {
			//fmt.Println("** end of tag", l.input[l.pos:])
			endTagCount--
			if endTagCount <= 0 {
				if l.pos > l.start {
					l.backup() // in questo punto si è posizionati sul primo carattere di endtokentag
					l.emit(itemTagChildContent)
					return nil
				}
				return l.errorf("Lex is wrong on tag %s", l.endtokentag)
			}
		}
		if l.next() == eof {
			return l.errorf("Malformed file, end of tag %s not found", l.endtokentag)
		}
	}
}

func lexStateTagName(l *lexer) stateFn {
	l.pos += len(l.tokentag)
	l.emit(itemTagName)
	return lexStateChild
}

func lexStateError(l *lexer) stateFn {
	return l.errorf("Template section not found")
}

func lexStateText(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], l.tokentag) {
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexStateTagName
		}
		if l.next() == eof {
			l.emit(itemTextWrong)
			return lexStateError
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
		state:       lexStateText,
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
		//fmt.Printf("*** type %v, val %q\n", item.typ, item.val)
		if item.typ == itemTextWrong {
			log.Println("This is a wrong template: ", item.val)
		}
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

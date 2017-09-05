package sqlread

import (
	"io"
)

type lexItemType uint8

//go:generate stringer -type=lexItemType
const (
	TIllegal lexItemType = iota

	TEof
	TSemi

	TComment

	TNull
	TString
	TNumber
	TIdentifier

	TDropTableFullStmt
	TLockTableFullStmt
	TUnlockTablesFullStmt

	TLParen
	TRParen

	TCreateTable
	TCreateTableDetail
	TCreateTableExtra

	TColumnType
	TColumnSize
	TColumnEnumVal
	TColumnDetails

	TInsertInto
	TInsertValues
	TInsertRow
)

type LexItem struct {
	Type lexItemType
	Val  string
	Pos  int64
}

type lexer struct {
	name  string
	input needToRead
	start int64
	pos   int64
	// width int
	items chan LexItem
}

const (
	eof  = byte(0)  // null
	lf   = byte(10) // \n
	semi = byte(59) // semicolon ;
	bs   = byte(92) // backslash \
	bt   = byte(96) // backtick `
	dot  = byte(46) // period .

	lprn = byte(40) // (
	rprn = byte(41) // )
	coma = byte(44) // ,

	sq = byte(39) // '
	dq = byte(34) // "

	letN = byte(78) // N
	// letn = byte(39)
)

func (l *lexer) next() byte {
	i, b := l.peak(1)
	if i != 1 {
		return eof
	}

	l.pos++

	return b[0]
}

func (l *lexer) rewind() {
	l.pos--
}

func (l *lexer) peak(s int) (int, []byte) {
	b := make([]byte, s)

	n, err := l.input.ReadAt(b, l.pos)
	if err != nil && err == io.EOF {
		panic(err)
		// return n, b
	}

	return n, b
}

func (l *lexer) hasPrefix(s string) bool {
	x := []byte(s)
	_, y := l.peak(len(x))

	return string(x) == string(y)
}

type needToRead interface {
	ReadAt(b []byte, off int64) (n int, err error)
}

type state func(*lexer) state

var (
	whitespace = []byte(" \t\r\n")
	sep        = []byte(" \t\r\n;")
	numbers    = []byte("0123456789")
)

func (l *lexer) Run() {
	for state := startState; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) accept(bs []byte) (c int) {
	for {
		n := l.next()

		found := false
		for _, b := range bs {
			if b == n {
				found = true
				break
			}
		}

		if !found {
			l.rewind()
			return c
		}
		c++
	}
}

func (l *lexer) until(b byte) bool {
	for {
		n := l.next()
		if n == eof {
			return false
		}

		if n == b {
			return true
		}
	}
}

func Lex(name string, input needToRead) (*lexer, chan LexItem) {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan LexItem),
	}

	return l, l.items
}

var item int64 = 0

func (l *lexer) emit(t lexItemType) LexItem {
	item++
	b := make([]byte, (l.pos - l.start))
	l.input.ReadAt(b, l.start)

	li := LexItem{
		Type: t,
		Val:  string(b),
		Pos:  l.start,
	}

	l.items <- li

	return li
}

func in(b byte, bs []byte) bool {
	for _, bb := range bs {
		if b == bb {
			return true
		}
	}

	return false
}

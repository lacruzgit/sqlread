package main

import (
	"github.com/chzyer/readline"
)

type ReadlineWrap struct {
	rl   *readline.Instance
	data []byte

	// start int64 coming optimization
	end int64
}

func NewReadlineWrap() *ReadlineWrap {
	rl, err := readline.NewEx(&readline.Config{
		Prompt: "> ",
		// HistoryFile:            "/tmp/readline-multiline",
		DisableAutoSaveHistory: true,
	})
	if err != nil {
		panic(err)
	}

	return &ReadlineWrap{
		rl:   rl,
		data: []byte{},
	}
}

func (s *ReadlineWrap) ReadAt(b []byte, off int64) (int, error) {
	e := off + int64(len(b)) - 1

	for e+1 > s.end {
		b2, err := s.rl.ReadSlice()
		b2 = append(b2, '\n')
		if err != nil {
			return 0, err // n value here is questionable
		}

		n := len(b2)
		s.end += int64(n)

		s.data = append(s.data, b2[:n]...)
	}

	i := 0
	for a := off; a <= e; a++ {
		b[i] = s.data[a]
		i++
	}

	return len(b), nil
}

func (s *ReadlineWrap) Flush() {
	s.end = 0
	s.data = []byte{}
}
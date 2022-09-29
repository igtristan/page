package main

import (
	"fmt"
	"io"
)

type Reader struct {
	index int
	data  []byte
}

func (r *Reader) LineCol(index int) (line int, col int) {

	for i := 0; i < index; i++ {
		if r.data[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	return line + 1, col + 1
}

func (r *Reader) ExpectString(str string) error {
	idx := r.index
	for _, c := range []byte(str) {
		if err := r.Expect(c); err != nil {
			line, col := r.LineCol(idx)
			return fmt.Errorf("%v:%v Unexpected string", line, col)
		}
	}
	return nil
}

func (r *Reader) Expect(b byte) error {
	pb, err := r.PeekByte()
	if err != nil {
		return err
	} else if pb != b {
		line, col := r.LineCol(r.index)
		return fmt.Errorf("%v:%v Unexpected character", line, col)
	}
	r.ReadByte()
	return nil
}

func (r *Reader) PeekByte() (byte, error) {
	if r.index >= len(r.data) {
		return 0, io.EOF
	}
	return r.data[r.index], nil
}

func (r *Reader) ReadByte() (byte, error) {
	if r.index >= len(r.data) {
		return 0, io.EOF
	}

	b := r.data[r.index]
	r.index++
	return b, nil
}

package main

import (
	"bytes"
	"fmt"
	"path/filepath"
)

type Scope struct {
	tags      map[string]*Tag
	FileScope *FileScope
}

func (s *Scope) CreateChild() *Scope {
	next := &Scope{
		tags: make(map[string]*Tag, len(s.tags)),
	}
	for k, v := range s.tags {
		next.tags[k] = v
	}
	next.FileScope = s.FileScope
	return next
}


func (s *Scope) CreateFileChild(path string) *Scope {
	tmp := s.CreateChild()
	tmp.FileScope = &FileScope{
		Path: path,
		Options: s.FileScope.Options,
		GlobalScope: s.GetGlobalScope(),
		UniqueClass: s.FileScope.UniqueClass,
	}
	return tmp
}

func (scope *Scope) GetGlobalScope() *GlobalScope{
	return scope.FileScope.GlobalScope
}

func (scope *Scope) GetFileScope() *FileScope{
	return scope.FileScope
}


///////////////////////////////////


type HtmlRenderingBuffer struct {
	Sequence int
	css      bytes.Buffer
}




type FileScope struct {

	Options  *processOptions

	// Globals
	GlobalScope *GlobalScope

	// Html rendering options
	UniqueClass *HtmlRenderingBuffer

	// Path that this file was found at
	Path        string
}

func (s *FileScope) NextClass() string {
	s.UniqueClass.Sequence++
	return fmt.Sprintf("p%v", s.UniqueClass.Sequence)
}

func (s *FileScope) AddCss(nodeType string, class string, subkey string, v string) {
	s.UniqueClass.css.WriteString(nodeType)
	s.UniqueClass.css.WriteByte('.')
	s.UniqueClass.css.WriteString(class)
	//s.css.WriteByte(' ')
	s.UniqueClass.css.WriteString(subkey)
	s.UniqueClass.css.WriteString(" {")
	s.UniqueClass.css.WriteString(v)
	s.UniqueClass.css.WriteString("}\n")
}

func (s *FileScope) AddMediaCss(m string, nodeType string, class string, subkey string, v string) {
	s.UniqueClass.css.WriteString("@media (" + m + "){")
	s.AddCss(nodeType, class, subkey, v)
	s.UniqueClass.css.WriteString("}")
}

func (s *FileScope) ResolvePath(path string) string {
	dir := filepath.Dir(s.Path)
	return filepath.Join(dir, path)
}

///////////////////////////////


type IncludeFile struct {
	tags *Tag
	path string
}

type GlobalScope struct {
	includes map[string]*IncludeFile
}

func NewGlobalScope() *GlobalScope {
	return &GlobalScope{
		includes: map[string]*IncludeFile{},
	}
}


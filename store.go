package tplexpr

import (
	"io"
	"io/fs"
	"sync"
	"time"
)

type simpleStore struct {
	c Context
}

var _ Store = &simpleStore{}

func (s *simpleStore) Render(w io.Writer, name string, vars Vars) error {
	return s.c.EvalTemplateWriter(name, vars, w)
}

type watchFile struct {
	fsys  fs.FS
	name  string
	mtime time.Time
}

type watchStore struct {
	mux         sync.Mutex
	plugins     []Plugin
	files       []storeFS
	parsed      bool
	c           Context
	watchFiles  []watchFile
	addBuiltins bool
}

var _ Store = &watchStore{}

func (s *watchStore) isExpired() bool {
	for _, wf := range s.watchFiles {
		s, err := fs.Stat(wf.fsys, wf.name)
		if err != nil {
			return true
		}
		if !s.ModTime().Equal(wf.mtime) {
			return true
		}
	}
	return false
}

func (s *watchStore) parseTemplate(name string, data []byte, cc *CompileContext) error {
	for _, p := range s.plugins {
		ok, err := p.ParseTemplate(name, data, cc)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
	return cc.ParseTemplate(name, data)
}

func (s *watchStore) parse() error {
	cc := NewCompileContext()
	s.watchFiles = s.watchFiles[:0]

	for _, f := range s.files {
		for _, glob := range f.globs {
			matches, err := fs.Glob(f.fs, glob)
			if err != nil {
				return err
			}

			for _, fileName := range matches {
				data, err := fs.ReadFile(f.fs, fileName)
				if err != nil {
					return err
				}
				st, err := fs.Stat(f.fs, fileName)
				if err != nil {
					return err
				}
				err = s.parseTemplate(fileName, data, &cc)
				if err != nil {
					return err
				}
				s.watchFiles = append(s.watchFiles, watchFile{f.fs, fileName, st.ModTime()})
			}
		}
	}

	s.parsed = true
	_, s.c = cc.Compile()
	if s.addBuiltins {
		AddBuiltins(&s.c)
	}
	for _, p := range s.plugins {
		p.InitContext(&s.c)
	}
	return nil
}

func (s *watchStore) updateWatchedFiles() error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if !s.parsed || s.isExpired() {
		s.parsed = false
		err := s.parse()
		return err
	}
	return nil
}

func (s *watchStore) Render(w io.Writer, name string, vars Vars) error {
	err := s.updateWatchedFiles()
	if err != nil {
		return err
	}
	return s.c.EvalTemplateWriter(name, vars, w)
}

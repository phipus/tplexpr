package tplexpr

import "io/fs"

type StoreBuilder struct {
	plugins []Plugin
	files   []storeFS
	watch   bool
}

type storeFS struct {
	fs    fs.FS
	globs []string
}

type Plugin interface {
	ParseTemplate(name string, data []byte, ctx *CompileContext) (compiled bool, err error)
}

func (s *StoreBuilder) AddPlugin(p Plugin) *StoreBuilder {
	s.plugins = append(s.plugins, p)
	return s
}

func (s *StoreBuilder) AddFS(fsys fs.FS, globs ...string) *StoreBuilder {
	s.files = append(s.files, storeFS{fsys, globs})
	return s
}

func (s *StoreBuilder) SetWatch(watch bool) *StoreBuilder {
	s.watch = watch
	return s
}

func (s *StoreBuilder) compileTemplate(name string, data []byte, cc *CompileContext) error {
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

func (s *StoreBuilder) Build() (Store, error) {
	if s.watch {
		return &watchStore{
			plugins: s.plugins,
			files:   s.files,
		}, nil
	}

	cc := NewCompileContext()
	for i := range s.files {
		f := &s.files[i]
		for _, g := range f.globs {
			matches, err := fs.Glob(f.fs, g)
			if err != nil {
				return nil, err
			}
			for _, fileName := range matches {
				data, err := fs.ReadFile(f.fs, fileName)
				if err != nil {
					return nil, err
				}
				err = s.compileTemplate(fileName, data, &cc)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	_, c := cc.Compile()
	return &simpleStore{c}, nil
}

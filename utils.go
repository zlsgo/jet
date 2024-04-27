package jet

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/zfile"
)

var (
	empty      = struct{}{}
	Extensions = []string{".jet", ".html"}
)

type Delims struct {
	Left  string
	Right string
}

type Options struct {
	Extensions []string
	Layout     string
	Debug      bool
	Reload     bool
	DelimLeft  string
	DelimRight string
}

func getOption(debug bool, opt ...func(o *Options)) Options {
	o := Options{
		Extensions: Extensions,
		DelimLeft:  "{{",
		DelimRight: "}}",
		Layout:     "slot",
	}
	if debug {
		o.Debug = true
		o.Reload = true
	}
	for _, f := range opt {
		f(&o)
	}
	return o
}

func ReadFile(path string, fs http.FileSystem) ([]byte, error) {
	if fs != nil {
		file, err := fs.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return io.ReadAll(file)
	}
	return zfile.ReadFile(path)
}

func (e *Engine) toName(path string) (name, rel string) {
	ext := filepath.Ext(path)
	if !zarray.Contains(e.options.Extensions, ext) {
		return
	}

	rel = zfile.SafePath(path, e.directory)
	name = strings.TrimSuffix(rel, ext)
	rel = strings.Replace(rel, "\\", "/", -1)

	return
}

func (e *Engine) readFile(path string) (buf []byte, err error) {
	if e.fileSystem != nil {
		var file http.File
		file, err = e.fileSystem.Open(path)
		if err != nil {
			return
		}
		defer file.Close()
		buf, err = io.ReadAll(file)
	} else {
		buf, err = zfile.ReadFile(path)
	}
	return
}

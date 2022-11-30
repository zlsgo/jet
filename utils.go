package jet

import (
	"io"
	"net/http"

	"github.com/sohaha/zlsgo/zfile"
)

var (
	empty      = struct{}{}
	extensions = []string{".html.jet", ".jet.html", ".jet"}
)

type Delims struct {
	Left  string
	Right string
}

type Options struct {
	Extension  string
	Layout     string
	Debug      bool
	Reload     bool
	DelimLeft  string
	DelimRight string
}

func getOption(debug bool, opt ...func(o *Options)) Options {
	o := Options{
		Extension:  ".jet.html",
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

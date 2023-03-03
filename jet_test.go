package jet

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zstring"
)

func TestRender(t *testing.T) {
	tt := zlsgo.NewTest(t)

	engine := New(nil, "./testdata/views", func(o *Options) {
		o.Debug = true
		o.Extensions = ".jet.html"
	})

	err := engine.Load()
	tt.NoError(err)

	var buf bytes.Buffer
	err = engine.Render(&buf, "index", map[string]interface{}{
		"Title": "Hello, World!",
	})
	tt.NoError(err)
	expect := `<h2>Header</h2><h1>Hello, World!</h1><h2>Footer</h2>`

	tt.Equal(expect, zstring.TrimLine(buf.String()))

	buf.Reset()

	err = engine.Render(&buf, "errors/404", map[string]interface{}{
		"Title": "Hello, World!",
	})
	tt.NoError(err)
	expect = `<h1>Hello, World!</h1>`
	tt.Equal(expect, zstring.TrimLine(buf.String()))
}

func TestLayout(t *testing.T) {
	tt := zlsgo.NewTest(t)

	engine := New(nil, "./testdata/views")

	err := engine.Load()
	tt.NoError(err)

	var buf bytes.Buffer
	err = engine.Render(&buf, "index", map[string]interface{}{
		"Title": "Hello, World!",
	}, "layouts/main")
	tt.NoError(err)

	expect := `<!DOCTYPE html><html><head><title>Title</title></head><body><h2>Header</h2><h1>Hello, World!</h1><h2>Footer</h2></body></html>`
	tt.Equal(expect, zstring.TrimLine(buf.String()))
}

func TestEmptyLayout(t *testing.T) {
	tt := zlsgo.NewTest(t)
	engine := New(nil, "./testdata/views")

	var buf bytes.Buffer

	err := engine.Render(&buf, "index", map[string]interface{}{
		"Title": "Hello, World!",
	}, "")
	tt.NoError(err)
	expect := `<h2>Header</h2><h1>Hello, World!</h1><h2>Footer</h2>`
	tt.Equal(expect, zstring.TrimLine(buf.String()))
}

func TestFileSystem(t *testing.T) {
	tt := zlsgo.NewTest(t)
	engine := NewFileSystem(nil, http.Dir(zfile.RealPath("./testdata/views")), func(o *Options) {
		o.Debug = true
	})

	var buf bytes.Buffer
	err := engine.Render(&buf, "index", map[string]interface{}{
		"Title": "Hello, World!",
	}, "/layouts/main")
	tt.NoError(err)

	expect := `<!DOCTYPE html><html><head><title>Title</title></head><body><h2>Header</h2><h1>Hello, World!</h1><h2>Footer</h2></body></html>`
	tt.Equal(expect, zstring.TrimLine(buf.String()))
}

func TestReload(t *testing.T) {
	tt := zlsgo.NewTest(t)
	engine := NewFileSystem(nil, http.Dir("./testdata/views"), func(o *Options) {
		o.Reload = true
	})

	err := engine.Load()
	tt.NoError(err)

	err = zfile.WriteFile("./testdata/views/reload.jet.html", []byte("after reload\n"))
	tt.NoError(err)

	defer func() {
		_ = zfile.WriteFile("./testdata/views/reload.jet.html", []byte("before reload\n"))
	}()

	_ = engine.Load()

	var buf bytes.Buffer
	err = engine.Render(&buf, "reload", nil)

	tt.NoError(err)
	expect := "after reload"
	tt.Equal(expect, zstring.TrimLine(buf.String()))
}

package jet

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/CloudyKit/jet/v6"
	"github.com/CloudyKit/jet/v6/loaders/httpfs"
	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
)

type Engine struct {
	directory  string
	log        *zlog.Logger
	fileSystem http.FileSystem
	loader     jet.Loader
	loaded     bool
	mutex      sync.RWMutex
	funcmap    map[string]interface{}
	Templates  *jet.Set
	options    Options
}

var _ znet.Template = &Engine{}

// New returns a Jet render engine for Fiber
func New(r *znet.Engine, directory string, opt ...func(o *Options)) *Engine {
	e := &Engine{
		directory: zfile.RealPath(directory),
		funcmap: map[string]interface{}{
			"toString": func(i interface{}) string {
				return ztype.ToString(i)
			},
			"toInt": func(i interface{}) int {
				return ztype.ToInt(i)
			},
		},
	}

	if r != nil {
		e.options = getOption(r.IsDebug(), opt...)
		e.log = r.Log
	} else {
		e.options = getOption(true, opt...)
		e.log = zlog.New()
		e.log.ResetFlags(zlog.BitLevel)
	}

	if !zarray.Contains(extensions, e.options.Extension) {
		e.log.Fatalf("%s extension is not a valid jet engine ['.html.jet', .jet.html', '.jet']", e.options.Extension)
	}

	return e
}

func NewFileSystem(r *znet.Engine, fs http.FileSystem, opt ...func(o *Options)) *Engine {
	e := New(r, "/", opt...)
	e.fileSystem = fs

	return e
}

// AddFunc adds the function to the template's function map
func (e *Engine) AddFunc(name string, fn interface{}) *Engine {
	e.mutex.Lock()
	e.funcmap[name] = fn
	e.mutex.Unlock()
	return e
}

// Exists returns whether or not a template exists under the requested path
func (e *Engine) Exists(templatePath string) bool {
	if !e.loaded || e.options.Reload {
		if err := e.Load(); err != nil {
			return false
		}
	}
	return e.loader.Exists(templatePath)
}

// Parse parses the templates to the engine
func (e *Engine) Load() (err error) {
	if e.loaded && !e.options.Reload {
		return nil
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.fileSystem != nil {
		e.loader, err = httpfs.NewLoader(e.fileSystem)
		if err != nil {
			return
		}
	} else {
		e.loader = jet.NewInMemLoader()
	}

	opts := []jet.Option{jet.WithDelims(e.options.DelimLeft, e.options.DelimRight)}

	if e.options.Debug {
		opts = append(opts, jet.InDevelopmentMode())

	}

	e.Templates = jet.NewSet(
		e.loader,
		opts...,
	)

	for name, fn := range e.funcmap {
		e.Templates.AddGlobal(name, fn)
	}

	if _, ok := e.loader.(*jet.InMemLoader); ok {
		total := 0
		tip := zstring.Buffer()
		err = filepath.Walk(e.directory, func(path string, info os.FileInfo, err error) error {
			l := e.loader.(*jet.InMemLoader)
			if err != nil {
				return err
			}
			if info == nil || info.IsDir() {
				return nil
			}
			if len(e.options.Extension) >= len(path) || path[len(path)-len(e.options.Extension):] != e.options.Extension {
				return nil
			}
			rel := zfile.SafePath(path, e.directory)
			name := strings.TrimSuffix(rel, e.options.Extension)
			name = strings.Replace(name, "\\", "/", -1)
			var buf []byte
			if e.fileSystem != nil {
				var file http.File
				file, err = e.fileSystem.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()
				buf, err = io.ReadAll(file)
			} else {
				buf, err = zfile.ReadFile(path)
			}
			if err != nil {
				return err
			}

			l.Set(name, string(buf))
			if e.options.Debug {
				total++
				tip.WriteString("\t    - " + name + "\n")
			}

			return err
		})

		if err == nil && !e.loaded && e.options.Debug {
			e.log.Debugf(zlog.ColorTextWrap(zlog.ColorLightGrey, "Loaded JET Templates (%d): \n%s"), total, tip.String())
		}

		e.loaded = true

	}

	return
}

// Execute will render the template by name
func (e *Engine) Render(out io.Writer, template string, data interface{}, layout ...string) error {
	if !e.loaded || e.options.Reload {
		if err := e.Load(); err != nil {
			return err
		}
	}
	tmpl, err := e.Templates.GetTemplate(template)
	if err != nil || tmpl == nil {
		return fmt.Errorf("render: template %s could not be loaded: %v", template, err)
	}

	var bind jet.VarMap
	if data != nil {
		if binds, ok := data.(map[string]interface{}); ok {
			bind = make(jet.VarMap)
			for key, value := range binds {
				bind.Set(key, value)
			}
		} else if binds, ok := data.(ztype.Map); ok {
			bind = make(jet.VarMap)
			for key, value := range binds {
				bind.Set(key, value)
			}
		} else if binds, ok := data.(jet.VarMap); ok {
			bind = binds
		}
	}

	if len(layout) > 0 && layout[0] != "" {
		lay, err := e.Templates.GetTemplate(layout[0])
		if err != nil {
			return err
		}

		bind.Set(e.options.Layout, func() {
			_ = tmpl.Execute(out, bind, empty)
		})
		return lay.Execute(out, bind, empty)
	}
	return tmpl.Execute(out, bind, empty)
}

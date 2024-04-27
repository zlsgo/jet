package jet

import (
	"sync"

	"github.com/CloudyKit/jet/v6"
	"github.com/fsnotify/fsnotify"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zstring"
)

var watcherOnce sync.Once

func changeHander(e *Engine) {
	if !e.options.Reload {
		return
	}

	watcherOnce.Do(func() {
		l, ok := e.loader.(*jet.InMemLoader)
		if !ok {
			return
		}

		if e.watcher != nil {
			go func() {
				for {
					select {
					case event, ok := <-e.watcher.Events:
						if !ok {
							return
						}

						name, rel := e.toName(event.Name)
						isDir := zfile.DirExist(event.Name)
						switch event.Op {
						case fsnotify.Create, fsnotify.Write:
							if isDir {
								e.watcher.Add(event.Name)
								continue
							}
							if name == "" {
								continue
							}
							buf, err := e.readFile(event.Name)
							if err != nil {
								e.log.Error("read file error", err)
								continue
							}
							l.Set(rel, zstring.Bytes2String(buf))
						case fsnotify.Remove, fsnotify.Rename:
							if isDir {
								e.watcher.Remove(event.Name)
								continue
							}
							if name == "" {
								continue
							}
							l.Delete(rel)
						}

					case err, ok := <-e.watcher.Errors:
						if !ok {
							return
						}
						e.log.Error("watcher error", err)
					}
				}
			}()
		}
	})
}

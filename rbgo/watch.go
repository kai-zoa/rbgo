package rbgo

import (
	"github.com/howeyc/fsnotify"
	"fmt"
	"path/filepath"
	"os"
	"time"
	"sync"
	"strings"
)

const (
	EventFound = EventName("Found")
	EventUpdate = EventName("Update")
	EventDelete = EventName("Delete")
)

type EventName string

type Event struct {
	Name    EventName
	Pacakge *Package
}

type EventBuffer struct {
	b []*Event
	m *sync.Mutex
	t time.Time
}

func (e *EventBuffer) init() {
	e.b = make([]*Event, 0)
	e.m = new(sync.Mutex)
	e.t = time.Now()
}

func (e *EventBuffer) add(event *Event) {
	e.m.Lock()
	defer e.m.Unlock()
	e.b = append(e.b, event)
	e.t = time.Now()
}

func (e *EventBuffer) fetch() (events []*Event) {
	e.m.Lock()
	defer e.m.Unlock()
	elapse := (time.Now().Sub(e.t) / time.Millisecond)
	if len(e.b) > 0 && elapse > 1000 {
		var prev *Event
		events = make([]*Event, 0, len(e.b))
		for _, ev := range e.b {
			if prev != nil &&
			prev.Name == ev.Name &&
			prev.Pacakge.WatchPath == ev.Pacakge.WatchPath {
				continue
			}
			events = append(events ,ev)
			prev = ev
		}
		e.b = make([]*Event, 0, len(e.b))
	}
	return events
}

type Watcher struct {
	Workspace *Workspace
	factory   *TaskFactory
}

func (w *Watcher) Watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	buf := EventBuffer{}
	buf.init()
	go func() {
		for {
			select {
			case fsev := <-watcher.Event:
				events := handleFSNotify(w.Workspace, fsev)
				for _, e := range events {
					if e.Name == EventFound {
						watcher.Watch(e.Pacakge.WatchPath)
					} else {
						buf.add(e)
					}
				}

			case event := <-watcher.Error:
				fmt.Println("error " + event.Error())
			}
		}
	}()

	n := 0
	vendorDir := filepath.Join(w.Workspace.sourceEntry, "vendor")
	err = w.Workspace.Walk(func(path string) error {
		if strings.HasPrefix(path, vendorDir) {
			return nil
		}
		n += 1
		return watcher.Watch(path)
	})
	fmt.Printf("Watch %d directories\n", n)
	if err != nil {
		return err
	}

	factory := TaskFactory{Package: w.Workspace.Package}
	runTask := func(pkg *Package) {
		task, err := factory.New(pkg.WatchPath)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		if err := build(task); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}

	// Build All
	fmt.Println("--- First Build Start")
	all := w.Workspace.Package.All()
	packages := make(map[string]*Package, len(all))
	for _, pkg := range all {
		packages[pkg.ObjectPath] = pkg
	}
	for _, pkg := range packages {
		runTask(pkg)
	}

	// Watch iNotify Events
	fmt.Println("--- Watch Start")
	for {
		if events := buf.fetch(); events != nil {
			for _, e := range events {
				fmt.Printf("%s: %s\n", e.Name, e.Pacakge.WatchPath)
				if e.Name == EventUpdate {
					runTask(e.Pacakge)
				}
			}
		}
		time.Sleep(time.Second)
	}
	return nil
}

func build(task *Task) error {
	for {
		if dep, err := task.FindDepends(); err != nil {
			return err
		} else if dep != nil {
			if err := build(dep); err != nil {
				return err
			}
		} else {
			break
		}
	}
	updated := false
	if fi, err := os.Stat(task.ObjectPath); err != nil {
		updated = true
	} else if (fi.ModTime().Before(task.Package.ModTime)) {
		updated = true
	}
	if updated {
		//fmt.Printf("Build: %s\n", task.ObjectPath)
		if err := task.Build(); err != nil {
			return err
		}
	}
	return nil
}

func handleFSNotify(ws *Workspace, event *fsnotify.FileEvent) []*Event {
	//fmt.Printf("%v\n", event)
	events := []*Event{} // FIXME
	defer func() {
		if len(events) > 0 {
			ws.Package.UpdateDepends()
		}
	}()
	path := event.Name
	fi, fsErr := os.Stat(event.Name)
	if fsErr != nil || fi == nil {
		if pkg := ws.Package.FindByPath(path); pkg != nil {
			ws.Package.Delete(pkg)
			events = append(events, &Event{Name: EventDelete, Pacakge: pkg})
		} else if pkg := ws.Package.FindByPath(filepath.Dir(path)); pkg != nil {
			pkg.Scan(ws.PackageRoot)
			events = append(events, &Event{Name: EventUpdate, Pacakge: pkg})
		}
		return events
	}
	if fi.IsDir() {

		ls := ws.Package.FindByDir(filepath.Dir(path))
		for _, pkg := range ls {
			if _, err := os.Stat(pkg.WatchPath); err != nil {
				ws.Package.Delete(pkg)
				events = append(events, &Event{Name: EventDelete, Pacakge: pkg})
			}
		}

	} else if IsGoSource(event.Name) {

		path = filepath.Dir(path)

	} else {
		return events
	}

	pkg := ws.Package.FindByPath(path)
	found := pkg != nil
	if !found {
		pkg = ws.NewPackage(path)
		events = append(events, &Event{Name: EventFound, Pacakge: pkg})
	}
	if err := pkg.Scan(ws.PackageRoot); err == SourceNotFound {
		if found {
			ws.Package.Delete(pkg)
			events = append(events, &Event{Name: EventDelete, Pacakge: pkg})
		}
	} else {
		//
		ws.Package.Put(pkg)
		events = append(events, &Event{Name: EventUpdate, Pacakge: pkg})
	}

	return events
}

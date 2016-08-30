package rbgo

import (
	"github.com/howeyc/fsnotify"
	"fmt"
	"path/filepath"
	"os"
	"time"
)

const (
	EventUpdate = EventName("Update")
	EventDelete = EventName("Delete")
)

type EventName string

type Event struct {
	Name    EventName
	Pacakge *Package
}

type Watcher struct {
	Workspace *Workspace
	event     chan *Event
	factory   *TaskFactory
}

func (w *Watcher) Watch() error {
	w.event = make(chan *Event)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	watchFn := func(path string) {
		watcher.Watch(path)
	}
	go func() {
		for {
			select {
			case event := <-watcher.Event:
				w.handleFSNotify(event, watchFn)

			case event := <-watcher.Error:
				fmt.Println("error " + event.Error())
			}
		}
	}()

	if err := w.Workspace.Walk(func(path string) error {
		return watcher.Watch(path)
	}); err != nil {
		return err
	}

	return w.watchEvents()
}

func (w *Watcher) handleFSNotify(event *fsnotify.FileEvent, watch func(string)) {
	fmt.Printf("%v\n", event)
	ws := w.Workspace
	events := []*Event{}
	defer func() {
		if len(events) > 0 {
			ws.Package.UpdateDepends()
		}
		for _, e := range events {
			w.event <- e
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
		return
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
		return
	}

	pkg := ws.Package.FindByPath(path)
	found := pkg != nil
	if !found {
		pkg = ws.NewPackage(path)
		watch(path)
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
}

func (w *Watcher) watchEvents() error {
	for {
		e := <-w.event
		time.Sleep(200 * time.Millisecond)
		receivedEvents := []*Event{e}
		if length := len(w.event); length > 0 {
			for i := 0; i < length; i += 1 {
				receivedEvents = append(receivedEvents, <-w.event)
			}
		}
		for _, e := range receivedEvents {

			fmt.Printf("%s: %s\n", e.Name, e.Pacakge.WatchPath)
		}
	}
	return nil
}

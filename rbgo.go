package main

import (
	. "./rbgo"
	"fmt"
)

func main() {
	ws, err := NewWorkspace(".")
	if err != nil {
		fmt.Println(err)
	}
	if err := ws.Init(); err != nil {
		fmt.Println(err)
	}
	watcher := Watcher{Workspace:ws}
	watcher.Watch()
}

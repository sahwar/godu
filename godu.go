package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/gdamore/tcell"
	"github.com/viktomas/godu/core"
)

func main() {
	limit := flag.Int64("l", 10, "show only files larger than limit (in MB)")
	flag.Parse()
	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}
	tree := core.GetSubTree(roots[0], ioutil.ReadDir, getIgnoredFolders())
	s := initScreen()
	commands := make(chan core.Executer)
	states := make(chan core.State)
	var wg sync.WaitGroup
	wg.Add(3)
	go core.StartProcessing(&tree, *limit*core.MEGABYTE, commands, states, &wg)
	go InteractiveTree(s, states, &wg)
	go ParseCommand(s, commands, &wg)
	wg.Wait()
	s.Fini()
}

func initScreen() tcell.Screen {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	if e = s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	s.Clear()
	return s
}

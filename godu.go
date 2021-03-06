package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/gdamore/tcell"
	"github.com/viktomas/godu/core"
	"github.com/viktomas/godu/interactive"
)

func main() {
	limit := flag.Int64("l", 10, "show only files larger than limit (in MB)")
	nullTerminate := flag.Bool("print0", false, "print null-terminated strings")
	flag.Parse()
	args := flag.Args()
	rootFolderName := "."
	if len(args) > 0 {
		rootFolderName = args[0]
	}
	rootFolderName, err := filepath.Abs(rootFolderName)
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Printf("godu will walk through `%s` that might take up to few minutes\n", rootFolderName)
	rootFolder := core.WalkFolder(rootFolderName, ioutil.ReadDir, getIgnoredFolders())
	rootFolder.Name = rootFolderName
	err = core.ProcessFolder(rootFolder, *limit*core.MEGABYTE)
	if err != nil {
		log.Fatalln(err.Error())
	}
	s := initScreen()
	commands := make(chan core.Executer)
	states := make(chan core.State)
	lastStateChan := make(chan *core.State, 1)
	var wg sync.WaitGroup
	wg.Add(3)
	go core.StartProcessing(rootFolder, commands, states, lastStateChan, &wg)
	go InteractiveFolder(s, states, &wg)
	go ParseCommand(s, commands, &wg)
	wg.Wait()
	s.Fini()
	lastState := <-lastStateChan
	printMarkedFiles(lastState, *nullTerminate)
}

func printMarkedFiles(lastState *core.State, nullTerminate bool) {
	markedFiles := interactive.FilesAsSlice(lastState.MarkedFiles)
	var printFunc func(string)
	if nullTerminate {
		printFunc = func(s string) {
			fmt.Printf("%s\x00", s)
		}
	} else {
		printFunc = func(s string) {
			fmt.Println(s)
		}
	}
	for _, f := range markedFiles {
		printFunc(f)
	}
}

func initScreen() tcell.Screen {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s, e := tcell.NewScreen()
	if e != nil {
		log.Printf("%v\n", e)
		os.Exit(1)
	}
	if e = s.Init(); e != nil {
		log.Printf("%v\n", e)
		os.Exit(1)
	}
	s.Clear()
	return s
}

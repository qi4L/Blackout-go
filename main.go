package main

import (
	"Blackout/Feature"
	"flag"
	"fmt"
)

var Process int

func usage() {
	fmt.Println(`Usage of main.exe:
  -p Process 
      指定进程
  `)
}

func main() {
	flag.IntVar(&Process, "u", 0, "target Process")
	flag.Usage = usage
	flag.Parse()

	Exp := Feature.WordExp{
		Process: Process,
		Drive:   "Blackout.sys",
	}
	Exp.Run()
}

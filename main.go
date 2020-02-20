package main

import (
	"fmt"
	"lmcli/processors"
	"lmcli/processors/createprocessing"
	"os"
)

const (
	StatusOk = iota
	StatusNoSuchCmd
)

var procs = map[string] processors.CmdProcessor {
	//"build" : buildprocessing.New(),
	//"b" : buildprocessing.New(),
	//"create": createprocessing.New(),
	//"c": createprocessing.New(),
	"cp": createprocessing.New(),
	//"cm": createprocessing.New(),
	//"test": testprocessing.New(),
	//"t": testprocessing.New(),
}

func welcomeMessage() {
	fmt.Println("LibMicroCMD LibMicro Command Line Interface")
	fmt.Println("Usage: lmcli <command> <subcommand> <parameter>\n")
	fmt.Println("  Commands:")
	//fmt.Println("\tbuild\t|\tb\t- builds project")
	fmt.Println("\tcp\t- creates a project")
	fmt.Println("\tcs\t- creates a service provider within project")
	fmt.Println("\t?\t- displays this help message")
	//fmt.Println("\ttest\t|\tt\t- performs project test")
	fmt.Println("")

	os.Exit(-1)
}

func executeProcessor(processors map[string]processors.CmdProcessor, args *[]string) uint8 {
	if prc,ok := processors[(*args)[0]]; ok  {
		prc.ProcessCmd(*args)
		return StatusOk
	}

	return StatusNoSuchCmd
}

func main() {
	gmc := os.Args[1:]

	if len(gmc) == 0 {
		welcomeMessage()
	}

	if executeProcessor(procs, &gmc) != StatusOk {
		welcomeMessage()
	}
}

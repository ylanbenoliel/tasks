package main

import (
	"flag"
	"fmt"
)

type CmdFlags struct {
	Add    string
	Del    int
	Toggle int
	List   bool
}

func NewCmdFlags() *CmdFlags {
	cf := CmdFlags{}

	flag.StringVar(&cf.Add, "add", "", "Add new task")
	flag.IntVar(&cf.Del, "del", -1, "Delete a task by index to delete")
	flag.IntVar(&cf.Toggle, "toggle", -1, "Specify a task by index to toggle")
	flag.BoolVar(&cf.List, "list", false, "List all tasks")

	flag.Parse()

	return &cf
}

func (cf *CmdFlags) Execute(tasks *Tasks) {
	switch {
	case cf.List:
		tasks.print()
	case cf.Del != -1:
		tasks.delete(cf.Del)
	case cf.Toggle != -1:
		tasks.toggle(cf.Toggle)
	case cf.Add != "":
		tasks.add(cf.Add)

	default:
		fmt.Println("Invalid command!")
	}

}

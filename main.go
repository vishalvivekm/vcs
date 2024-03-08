package main

import (
	"bufio"
	"fmt"
	_"io"
	"log"
	"os"
	_"reflect"
)

type fullHelpText string

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	helptxt := fullHelpText(`These are SVCS commands:
config     Get and set a username.
add        Add a file to the index.
log        Show commit logs.
commit     Save changes.
checkout   Restore a file.`)
	if len(os.Args) < 2 {
		fmt.Println("Uh-Oh! looks you've not passed any arguments.", printFullHelp(helptxt))
		log.Fatal("Please choose form any one of these available commands :)")
	}
	if len(os.Args) == 3 {
		log.Fatal("Please pass only one command at a time ;)")
	}
	
	cmd := os.Args[1]

	switch cmd {
	case "--help":
		printFullHelp(helptxt)
	case "config", "log ":
		fmt.Println("Get and log and set a username.")
	case "add":
		fmt.Println("Add a file to the index.")
	case "log":
		fmt.Println("Show commit logs.")
	case "commit":
		fmt.Println("Save changes.")
	case "checkout":
		fmt.Println("Restore a file.")
	default:
		fmt.Printf("%s is not one of the available commands\nList avilable commands: y or n?", cmd)
		scanner.Scan()
		if scanner.Text() == "y" {
			fmt.Println(printFullHelp(helptxt))
		} else {
			fmt.Println("Bye")
		}

	}
}

func printFullHelp(helptxt fullHelpText) string {
	return fmt.Sprintf("%s", helptxt)
}


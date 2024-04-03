package main

import (
	"fmt"
	"os"
	"strings"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {

	Args := os.Args
	if len(os.Args) < 2 {
		// print all the commands
		fmt.Println(helptxt)
		return
	}

	err := os.Mkdir("vcs", 0755)
	if err != nil && os.IsExist(err) {
		// do nothing
	} else {
		check(err)
	}

	createFile := func(name string, data string) {
		d := []byte(data)
		check(os.WriteFile(name, d, 0644))
	}

	_, err = os.Stat("vcs/config.txt")
	if err != nil && os.IsNotExist(err) {
		createFile("vcs/config.txt", "")
	} else {
		check(err)
	}
	_, err = os.Stat("vcs/index.txt")
	if err != nil && os.IsNotExist(err) {
		createFile("vcs/index.txt", "")
	} else {
		check(err)
	}

	switch Args[1] {
	case "config":
		if len(Args) < 3 { // .main config
			// just read the vcs/confix.txt and print that

			content := string(readFileContent("vcs/config.txt"))
			if len(content) == 0 {
				fmt.Println("Please, tell me who you are")
				return
			}
			fmt.Printf("The username is %s\n", content)
			return
		} else {
			// write the username Args[2] to vcs/config.txt and read -> output that
			err := os.WriteFile("vcs/config.txt", []byte(Args[2]), 0755)
			check(err)
			content := readFileContent("vcs/config.txt")
			fmt.Printf("The username is %s\n", string(content))
			return
		}
	case "add":
		if len(Args) < 3 {
			// print all the files being tracked
			content := string(readFileContent("vcs/index.txt"))
			if len(content) == 0 {
				fmt.Println("Add a file to the index")
				return
			}
			fmt.Printf("Tracked files:\n%s", content)
			return
		} else {
			fileNames := Args[2:]
			// check if the file is already being tracked
			filesdata := string(readFileContent("vcs/index.txt"))
			s := strings.Contains(filesdata, Args[2])
			if s {
				fmt.Printf("The file %s is already being tracked\n", Args[2])
				return
			}

			data := []string{}
			for _, filename := range fileNames {
				file, err := os.Stat(filename)
				if err != nil {
					if os.IsNotExist(err) {
						fmt.Printf("can not find %s\n", filename)
						continue
					} else {
						check(err)
					}
				}
				data = append(data, file.Name())

			}
			//	fmt.Println(data)
			file, err := os.OpenFile("vcs/index.txt", os.O_APPEND|os.O_WRONLY, 0644)
			check(err)
			defer file.Close()

			for i, fname := range data {
				_, err = fmt.Fprintln(file, fname) // writes each file's name\n to the next line
				if err != nil {
					check(err)
				}

				if i == len(data)-1 {
					fmt.Printf("Tracking:\n%s", string(readFileContent("vcs/index.txt"))) // output files being tracked
				}
			}

		}
	default:
		fmt.Printf("command: %s not supported yet\n\nAvailable:\n%s\n", Args[1], helptxt)
	}

}
func readFileContent(name string) []byte {

	data, err := os.ReadFile(name)
	check(err)
	return data

	//file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0755)
	//check(err)
	//defer file.Close()
	//buf := make([]byte, 1024)
	//n, err = file.Read(buf)
	//check(err)

}

var helptxt = `These are SVCS commands:
config     Get and set a username.
add        Add a file to the index.`

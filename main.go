package main

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

var ch = make(chan string)
var wg sync.WaitGroup

func check(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	Args := os.Args
	if len(os.Args) < 2 {
		// print help
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
			// just read the vcs/config.txt and print that

			content := string(readFileContent("vcs/config.txt"))
			if len(content) == 0 {
				fmt.Println("Please, set a username")
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
	case "log":
		_, err := os.Stat("vcs/log.txt")
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("No commits made yet")
			return
		}
		file, err := os.Open("vcs/log.txt")
		check(err)

		fmt.Println(strings.Trim(string(readFileContent(file.Name())), "\n"))
	case "commit": // ./main commit "msg" 0 1 2
		author := readFileContent("vcs/config.txt")
		if string(author) == "" {
			fmt.Println("Please configure Username first") // can prompt for help ? y or n : showHelp
			return
		}
		content := string(readFileContent("vcs/index.txt"))
		if len(content) == 0 {
			fmt.Println("No tracked files, add a file to the index first")
			return
		}
		if len(Args) < 3 {
			fmt.Println("Please make sure to pass the commit msg")
			return
		}

		cm := commitObject{
			Author: string(author),
			Msg:    Args[2],
		}
		content = string(readFileContent("vcs/index.txt"))
		if len(content) == 0 {
			fmt.Println("No tracked files, add a file to the index first")
			return
		}
		err = os.MkdirAll("vcs/commits", os.ModePerm)
		check(err)
		var hashesOfFiles []string
		// now read all the file being tracked, create their individual hashes, and hash the hashes and try to cretae a directory
		// if errors.Is(err, os.IsExist) - no changes have been made,return
		// copy the files to the /commits/newHash/
		file, err := os.Open("vcs/index.txt")
		check(err)
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			filebeingTracked := scanner.Text()
			content := readFileContent(filebeingTracked)
			hash := sha256.New()
			_, err = hash.Write(content)
			check(err)
			fileHash := fmt.Sprintf("%x", hash.Sum(nil))
			hashesOfFiles = append(hashesOfFiles, fileHash)
		}
		combinedHashOfFiles := strings.Join(hashesOfFiles, "")
		hash := sha256.New()
		_, err = hash.Write([]byte(combinedHashOfFiles))
		check(err)
		commitHash := fmt.Sprintf("%x", hash.Sum(nil))
		/*fmt.Println("Here is the hash of all the files combined: ")
		fmt.Println(commitHash)*/
		err = os.Mkdir(fmt.Sprintf("vcs/commits/%s", commitHash), os.ModePerm)
		if err != nil {
			if os.IsExist(err) {
				fmt.Println("No Files were changed")
				return
			}
			fmt.Println("can not create dir: ")
			log.Fatalln(err)
		}
		cm.Commit = commitHash
		logMsg := fmt.Sprintf("commit %s\nAuthor: %s\n%s", cm.Commit, cm.Author, cm.Msg)
		//fmt.Println(logMsg)

		// write logs to log.txt
		file, err = os.OpenFile("vcs/log.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
		check(err)
		defer file.Close()
		content = string(readFileContent(file.Name()))
		err = os.WriteFile(file.Name(), []byte(logMsg), os.ModePerm)
		check(err)
		_, err = fmt.Fprint(file, "\n\n", content)
		check(err)

		// copy files of current commit to vcs/commits/commitID

		//file, err = os.Open("vcs/index.txt")
		//if err != nil {
		//	log.Fatalln(err)
		//}
		//defer file.Close()
		//scanner = bufio.NewScanner(file)
		//for scanner.Scan() {
		//	if singleFile := scanner.Text(); len(singleFile) != 0 {
		//		//command := fmt.Sprintf("cp %s", singleFile)
		//		cmd := exec.Command("cp", singleFile, fmt.Sprintf("vcs/commits/%s", cm.Commit))
		//		if err := cmd.Run(); err != nil {
		//			log.Fatalf("can not copy files to commits: %s", err)
		//		}
		//	}
		//
		//}
		wg.Add(2)
		go readIndex() // // reads the file being tracked and sends filePaths to the channel
		go readFilesAndCopy(cm.Commit)

	default:
		fmt.Printf("command: %s not supported yet\n\nAvailable:\n%s\n", Args[1], helptxt)
	}

	wg.Wait()
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
config     Get and set a username
add        Add a file to the index
log        Show commit logs
commit     Save changes`

type commitObject struct {
	Author, Commit, Msg string
}

func readIndex() {
	defer wg.Done()
	file, err := os.Open("vcs/index.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		singleFile := scanner.Text()
		if len(singleFile) != 0 {
			ch <- singleFile
		}
	}
	close(ch)
}

func readFilesAndCopy(commitHash string) {
    defer wg.Done()
    for fileName := range ch {
        content := readFileContent(fileName)
        err := os.WriteFile(fmt.Sprintf("vcs/commits/%s/%s", commitHash, fileName), content, os.ModePerm)
        if err != nil {
            log.Fatalf("Error writing file %s to commit: %v", fileName, err)
        }
        fmt.Printf("Copied file:%s to the commit %s\n", fileName, commitHash)
    }
}

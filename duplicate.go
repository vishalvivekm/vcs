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
	"vishalvivekm/vcs/utils"
)

var ch = make(chan string)
var wg sync.WaitGroup

func main() {

	Args := os.Args
	if len(os.Args) < 2 {
		// print help
		fmt.Println(helpTxt)
		return
	}

	err := os.Mkdir("vcs", 0755)
	if err != nil && os.IsExist(err) {
		// do nothing
	} else {
		utils.Check(fmt.Errorf("error while creating vcs directory: %w", err))
	}

	createFile := func(name string, data string) {
		d := []byte(data)
		utils.Check(os.WriteFile(name, d, 0644))
	}

	_, err = os.Stat(configFile)
	if err != nil && os.IsNotExist(err) {
		utils.Check(utils.CreateFile(configFile, ""))
	} else {
		utils.Check(err)
	}
	_, err = os.Stat(indexFile)
	if err != nil && os.IsNotExist(err) {
		createFile(indexFile, "")
	} else {
		utils.Check(err)
	}

	switch Args[1] {
	case "config":
		configUser(Args)
	case "add":
		addFilesToIndex(Args)

	case "log":
		displayLogs()
	case "commit":
		author := readFileContent(configFile)
		if string(author) == "" {
			fmt.Println("Please configure Username first") // can prompt for help ? y or n : showHelp
			return
		}
		content := string(readFileContent(indexFile))
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
		content = string(readFileContent(indexFile))
		if len(content) == 0 {
			fmt.Println("No tracked files, add a file to the index first")
			return
		}
		err = os.MkdirAll(commitsDir, os.ModePerm)
		utils.Check(err)
		var hashesOfFiles []string

		// now read all the file being tracked, hash them,
		//then hash the hashes to get a unique commitId
		file, err := os.Open(indexFile)
		utils.Check(err)
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			fileBeingTracked := scanner.Text()
			content := readFileContent(fileBeingTracked)
			fileHash, err := returnHash(content)
			utils.Check(err)
			hashesOfFiles = append(hashesOfFiles, fileHash)
		}
		if err := scanner.Err(); err != nil {
			utils.Check(err)
		}
		combinedHashOfFiles := strings.Join(hashesOfFiles, "")
		commitHash, err := returnHash([]byte(combinedHashOfFiles))
		utils.Check(err)
		/*fmt.Println("Here is the hash of all the files combined: ")
		fmt.Println(commitHash)*/

		// compare changes from previous commit
		cm.Commit = commitHash
		isHashSameAsPrevCommitId, err := utils.CompareCommit(cm.Commit)
		if err != nil {
			utils.Check(err)
		}
		if isHashSameAsPrevCommitId {
			fmt.Println("No Files were changed")
			return
		}
		err = os.Mkdir(fmt.Sprintf("%s/%s", commitsDir, cm.Commit), os.ModePerm)
		if err != nil {
			if os.IsExist(err) { //or, if errors.Is(err, os.IsExist)

				// do nothing: refer issue: https://github.com/vishalvivekm/vcs/issues/3

			} else {
				utils.Check(err)
			}
		}
		logMsg := fmt.Sprintf("commit %s\nAuthor: %s\n%s", cm.Commit, cm.Author, cm.Msg)

		// write logs to logFile
		utils.Check(writeLogs(logMsg))

		wg.Add(2)
		go readIndex()                 // reads the file being tracked and sends filePaths to the channel ch
		go readFilesAndCopy(cm.Commit) // reads files at paths received from channel and copies them to
		// vcs/commits/{commitID}

	default:
		fmt.Printf("command: %s not supported yet\n\nAvailable:\n%s\n", Args[1], helpTxt)
	}
	wg.Wait()
}

func readFileContent(name string) []byte {

	data, err := os.ReadFile(name)
	utils.Check(err)
	return data

}

var (
	logFile    = "vcs/log.txt"
	indexFile  = "vcs/index.txt"
	configFile = "vcs/config.txt"
	commitsDir = "vcs/commits"
)
var helpTxt = `These are available commands:
config     Get and set a username
add        Add a file to the index
log        Show commit logs
commit     Save changes`

type commitObject struct {
	Author, Commit, Msg string
}

func readIndex() {
	defer wg.Done()
	file, err := os.Open(indexFile)
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
	if err := scanner.Err(); err != nil {
		utils.Check(err)
	}
	close(ch)
}

func readFilesAndCopy(commitHash string) {
	defer wg.Done()
	for fileName := range ch {
		content := readFileContent(fileName)
		err := os.WriteFile(fmt.Sprintf("%s/%s/%s", commitsDir, commitHash, fileName), content, os.ModePerm)
		if err != nil {
			log.Fatalf("Error writing file %s to commit: %v", fileName, err)
		}
		fmt.Printf("Copied file:%s to the commit %s\n", fileName, commitHash)
	}
}

func returnHash(content []byte) (string, error) {
	hash := sha256.New()
	_, err := hash.Write(content)
	if err != nil {
		return "", err
	}
	hashedString := fmt.Sprintf("%x", hash.Sum(nil))
	return hashedString, nil
}

func configUser(cliArgs []string) {
	if len(cliArgs) < 3 { // .main config
		// just read the configFile and print that
		content := string(readFileContent(configFile))
		if len(content) == 0 {
			fmt.Println("Please, set a username")
			return
		}
		fmt.Printf("The username is %s\n", content)
		return
	} else {
		// write the username cliArgs[2] to configFile and output that
		err := os.WriteFile(configFile, []byte(cliArgs[2]), 0755)
		utils.Check(err)
		content := readFileContent(configFile)
		fmt.Printf("The username is %s\n", string(content))
		return
	}
}

func addFilesToIndex(cliArgs []string) {
	if len(cliArgs) < 3 { // ./main add
		// print all the files being tracked
		content := string(readFileContent(indexFile))
		if len(content) == 0 { // no files are being tracked
			fmt.Println("Add a file to the index")
			return
		}
		fmt.Printf("Tracked files:\n%s", content)
		return
	} else { // .main add <filenames>...
		fileNames := cliArgs[2:]
		// check if the file is already being tracked
		trackedFiles := string(readFileContent(indexFile))
		doesFileExistInTrackingIndex := strings.Contains(trackedFiles, cliArgs[2])
		if doesFileExistInTrackingIndex {
			fmt.Printf("The file %s is already being tracked\n", cliArgs[2])
			return
		}
		var filesToAddInTrackingIndex []string
		for _, filename := range fileNames {
			file, err := os.Stat(filename)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Printf("can not find %s\n", filename)
					continue
				} else {
					utils.Check(err)
				}
			}
			filesToAddInTrackingIndex = append(filesToAddInTrackingIndex, file.Name())

		}
		file, err := os.OpenFile(indexFile, os.O_APPEND|os.O_WRONLY, 0644)
		utils.Check(err)
		defer file.Close()

		for i, fName := range filesToAddInTrackingIndex {
			// writes each file's name\n onto the next line
			_, err = fmt.Fprintln(file, fName)
			if err != nil {
				utils.Check(err)
			}

			if i == len(filesToAddInTrackingIndex)-1 {
				// display files being tracked
				fmt.Printf("Tracking:\n%s", string(readFileContent(indexFile)))
			}
		}

	}
}

func displayLogs() {
	_, err := os.Stat(logFile)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("No commits made yet")
		return
	}
	file, err := os.Open(logFile)
	utils.Check(err)

	fmt.Println(strings.Trim(string(readFileContent(file.Name())), "\n"))
}
func writeLogs(logMsg string) error {
	file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	content := string(readFileContent(file.Name()))
	if err := os.WriteFile(file.Name(), []byte(logMsg), os.ModePerm); err != nil {
		return err
	}
	if _, err = fmt.Fprint(file, "\n\n", content); err != nil {
		return err
	}
	return nil
}

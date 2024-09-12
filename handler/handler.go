package handler

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	config "vishalvivekm/vcs/constants"
	"vishalvivekm/vcs/types"
	"vishalvivekm/vcs/utils"
	"sync"
)
var ch = make(chan string)
var wg sync.WaitGroup

func ConfigUser(cliArgs []string) {
	if len(cliArgs) < 3 { // .main config
		// just read the configFile and print that
		content, err := utils.ReadFileContent(config.ConfigFile)
		if err != nil {
			utils.Check(err)
		}
		if len(content) == 0 {
			fmt.Println("Please, set a username")
			return
		}
		fmt.Printf("The username is %s\n", string(content))
		return
	} else {
		// write the username cliArgs[2] to configFile and output that
		err := os.WriteFile(config.ConfigFile, []byte(cliArgs[2]), 0755)
		utils.Check(err)
		content, err := utils.ReadFileContent(config.ConfigFile)
		utils.Check(err)
		fmt.Printf("The username is %s\n", string(content))
		return
	}
}

func AddFilesToIndex(cliArgs []string) {
	if len(cliArgs) < 3 { // ./main add
		// print all the files being tracked
		content, err := utils.ReadFileContent(config.IndexFile)
		if err != nil {
			utils.Check(err)
		}
		if len(content) == 0 { // no files are being tracked
			fmt.Println("Add a file to the index")
			return
		}
		fmt.Printf("Tracked files:\n%s", string(content))
		return
	} else { // .main add <filenames>...
		fileNames := cliArgs[2:]
		// check if the file is already being tracked
		indexedFiles, err := utils.ReadFileContent(config.IndexFile)
		if err != nil {
			utils.Check(err)
		}
		trackedFiles := string(indexedFiles)
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
		file, err := os.OpenFile(config.IndexFile, os.O_APPEND|os.O_WRONLY, 0644)
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
				content, err := utils.ReadFileContent(config.IndexFile)
				utils.Check(err)
				trackedFiles = string(content)
				fmt.Printf("Tracking:\n%s", trackedFiles)
			}
		}

	}
}

func DisplayLogs() {
	_, err := os.Stat(config.LogFile)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("No commits made yet")
		return
	}
	file, err := os.Open(config.LogFile)
	utils.Check(err)
	logs, err := utils.ReadFileContent(file.Name())
	if err != nil {
		utils.Check(err)
	}
	fmt.Println(strings.Trim(string(logs), "\n"))
}

func Commit(cm *types.CommitObject) {
err := os.MkdirAll(config.CommitsDir, os.ModePerm)
		utils.Check(err)
		var hashesOfFiles []string

		// now read all the file being tracked, hash them,
		//then hash the hashes to get a unique commitId
		file, err := os.Open(config.IndexFile)
		utils.Check(err)
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			fileBeingTracked := scanner.Text()
			content, err := utils.ReadFileContent(fileBeingTracked)
			utils.Check(err)
			fileHash, err := utils.ReturnHash(content)
			utils.Check(err)
			hashesOfFiles = append(hashesOfFiles, fileHash)
		}
		if err := scanner.Err(); err != nil {
			utils.Check(err)
		}
		combinedHashOfFiles := strings.Join(hashesOfFiles, "")
		commitHash, err := utils.ReturnHash([]byte(combinedHashOfFiles))
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
		err = os.Mkdir(fmt.Sprintf("%s/%s", config.CommitsDir, cm.Commit), os.ModePerm)
		if err != nil {
			if os.IsExist(err) { //or, if errors.Is(err, os.IsExist)

				// do nothing: refer issue: https://github.com/vishalvivekm/vcs/issues/3

			} else {
				utils.Check(err)
			}
		}
		logMsg := fmt.Sprintf("commit %s\nAuthor: %s\n%s\nTime: %s", cm.Commit, cm.Author, cm.Msg, cm.Date)

		// write logs to logFile
		utils.Check(utils.WriteLogs(logMsg))

		wg.Add(2)
		go utils.ReadIndex(&wg, ch)        // reads the file being tracked and sends filePaths to the channel ch
		go utils.ReadFilesAndCopy(&wg, ch, cm.Commit) // reads files at paths received from channel and copies them to
		// vcs/commits/{commitID}
		wg.Wait()
}


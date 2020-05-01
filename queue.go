package queue

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var (
	wg sync.WaitGroup
)

//Element :- struct formater
type Element struct {
	data    string
	retries int
}

//Config :- struct formater
type Config struct {
	queue      []*Element
	isWritting bool
	isReading  bool
	mu         sync.Mutex
	//user
	StoragePath string
	FileSize    int64
	NoOfRetries int
}

//Init :- main function of module
func (c *Config) Init() {
	c.createTempFolderIfNotExist()
	go c.scheduleQueueCheck(30 * time.Second)
}

//Insert : - export function of module
func (c *Config) Insert(data string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ele := new(Element)
	ele.data = data
	ele.retries = 0
	c.queue = append(c.queue, ele)
}

func (c *Config) scheduleQueueCheck(interval time.Duration) {

	ticker := time.NewTicker(interval)
	for range ticker.C {
		if !c.isWritting {
			c.processQueue()
		}
	}

}

func (c *Config) processQueue() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.isWritting = true
	for len(c.queue) > 0 {
		msg := c.queue[0]

		if c.writeToFile(msg.data) {
			c.queue = c.queue[1:]
			continue
		}

		if c.queue[0].retries >= c.NoOfRetries {
			c.queue = c.queue[1:]
		}
		c.queue[0].retries++
	}
	c.isWritting = false
}

//Writre to file logic
func (c *Config) writeToFile(data string) bool {
	// file := createFileIfNotExist("just.txt")
	// if getFileSize(file) > c.FileSize {
	// 	file.Close()
	// 	c.createTempFolderIfNotExist()
	// 	c.moveFileToTemp()
	// 	file = createFileIfNotExist("just.txt")
	// }
	// defer file.Close()
	// status := appendToFile(file, data)
	// return status
	file := createFileIfNotExist("just.txt")
	defer file.Close()

	status := appendToFile(file, data)

	if getFileSize(file) > c.FileSize {
		c.createTempFolderIfNotExist()
		c.moveFileToTemp()
	}

	return status
}

func (c *Config) createTempFolderIfNotExist() {
	os.Mkdir(c.StoragePath, 0755)
}

func (c *Config) moveFileToTemp() { //Readfiles from here
	renameError := os.Rename("just.txt", getFilePath(c.StoragePath))
	LogError("Cannot rename/move file", renameError)
}

func appendToFile(file *os.File, data string) bool {
	_, appendToFileError := file.WriteString(data)
	if appendToFileError != nil {
		LogError("Cannot append data to file", appendToFileError)
		return false
	} else {
		return true
	}
}

func createFileIfNotExist(fileName string) *os.File {
	file, fileOpenError := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	LogError("Cannot open file", fileOpenError)
	return file
}

func getFileSize(file *os.File) int64 {
	fileInfo, fileInfoError := file.Stat()
	if fileInfoError != nil {
		LogError("Unable to get file size", fileInfoError)
		return 0
	}

	const mb = 1024 * 1024 
	return fileInfo.Size() / mb
}

func getFilePath(path string) string {
	return path + "/" + strconv.Itoa(rand.Intn(6645600532184904)) + "_" + time.Now().String() + ".txt"
}

// Write to file logic end

//Read from file logic start

func (c *Config) Read(callback func(message string, fileName string)) {
	c.schedule(callback, 30*time.Second)
}

func (c *Config) schedule(callback func(message string, fileName string), interval time.Duration) {
	fmt.Println("Reading file status**:", c.isReading)

	ticker := time.NewTicker(interval)
	for range ticker.C {

		if !c.isReading {
			c.isReading = true
			var files = c.getAllFileFromDir()

			if files != nil && len(files) > 0 {
				fmt.Println("files available to read:", files, "Reading file status**:", c.isReading)
				for _, file := range files {
					fmt.Println(file.Mode().IsRegular(), filepath.Ext(file.Name()))
					if file.Mode().IsRegular() && filepath.Ext(file.Name()) == ".txt" {
						fileData := c.readFile(file.Name())
						fileInfo := convertFileDataToString(fileData)
						callback(fileInfo, file.Name())
					}
				}
				fmt.Println("Finished reading file status**:", c.isReading)
			}
			c.isReading = false //called once for loop is completed
		}

	}
}

//CommitFile :- removes file from storage
func (c *Config) CommitFile(fileName string) {
	c.removeFile(fileName)
}

func (c *Config) getAllFileFromDir() []os.FileInfo {
	files, filesInfoError := ioutil.ReadDir(c.StoragePath)
	LogError("Unable to get files form dir", filesInfoError)
	return files
}

func (c *Config) readFile(fileName string) []byte {
	fileData, fileReadError := ioutil.ReadFile(c.StoragePath + "/" + fileName) // just pass the file name
	LogError("Unable to read from file", fileReadError)
	return fileData
}

func convertFileDataToString(fileData []byte) string {
	return string(fileData)
}

func (c *Config) removeFile(fileName string) {
	removeFileError := os.Remove(c.StoragePath + "/" + fileName)
	fmt.Println("fileremoved: ", fileName)
	LogError("Unable to delete file", removeFileError)
}

//LogError simple error log
func LogError(message string, errorData error) {
	if errorData != nil {
		fmt.Println("Error from queue.go : -> "+message, errorData)
	}
}

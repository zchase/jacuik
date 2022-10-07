package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

func CreateDirectory(dirPath string) error {
	return os.Mkdir(dirPath, os.ModePerm)
}

func WriteFile(filePath, contents string) error {
	return ioutil.WriteFile(filePath, []byte(contents), 0644)
}

func WriteYAMLFile(fileName string, data interface{}) error {
	yamlData, err := yaml.Marshal(&data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, yamlData, 0644)
}

func WriteJSONFile(fileName string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, jsonData, 0644)
}

func ParseYAMLFile[T any](filePath string) (*T, error) {
	result := new(T)

	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(fileContents, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func ParseJSONFile[T any](filePath string) (*T, error) {
	result := new(T)

	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(fileContents, result)
	if err != nil {
		return result, err
	}

	return result, nil
}

// ReadDirectoryContents reads the contents of a directory and returns
// a list of the names of the contents.
func ReadDirectoryContents(dirPath string) ([]string, error) {
	directory, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}

	contents, err := directory.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

// GetWorkingDirectoryName returns the name of the working directory.
func GetWorkingDirectoryName() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if wd == "" {
		return "", fmt.Errorf("Could not find working directory.")
	}

	wdParts := strings.Split(wd, "/")
	wdName := wdParts[len(wdParts)-1]
	return wdName, nil
}

// IsCurrentDirectoryEmpty checks to see if the current working
// directory is empty.
func IsCurrentDirectoryEmpty() (bool, error) {
	// Get the current working directory.
	path, err := os.Getwd()
	if err != nil {
		return false, err
	}

	// Open the current working directory.
	directory, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer directory.Close()

	// Check if the directory is empty.
	c, err := directory.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}

	fmt.Println(c)

	// The directory is either not empty or
	// has returned an error.
	return false, err
}

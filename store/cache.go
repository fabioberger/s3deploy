package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/go-errors/errors"
	homedir "github.com/mitchellh/go-homedir"
)

// Cache stores which files have been seen before and a hash of their contents during it's previous
// upload to S3
type Cache struct {
	filePathToHash map[string]string
}

// NewCache instantiates a new instance of the Cache
func NewCache() (*Cache, error) {
	if err := createCacheFileInNonExists(); err != nil {
		return nil, err
	}

	filePathToHash, err := getFilePathToHashFromFile()
	if err != nil {
		return nil, err
	}
	return &Cache{
		filePathToHash: filePathToHash,
	}, nil
}

// Get retrieves a hash given a filePath
func (c *Cache) Get(filePath string) (hash string, ok bool) {
	hash, ok = c.filePathToHash[filePath]
	return hash, ok
}

// Set updates the hash for a given filePath in the cache
func (c *Cache) Set(filePath, hash string) error {
	r, err := regexp.Compile("[A-Z]")
	if err != nil {
		return err
	}
	if len(hash) != 64 || r.MatchString(hash) {
		return errors.Errorf(`Attempted to set cache entry for filePath: %s with invalid hash: %s.
            Hash must be sha256, base-16 encoded with lowercase letters.`, filePath, hash)
	}
	c.filePathToHash[filePath] = hash
	return nil
}

// Remove deletes an entry from the cache
func (c *Cache) Remove(filePath string) {
	delete(c.filePathToHash, filePath)
}

// GetFilePaths returns the filePaths saved in the cache
func (c *Cache) GetFilePaths() []string {
	filePaths := []string{}
	for filePath := range c.filePathToHash {
		filePaths = append(filePaths, filePath)
	}
	return filePaths
}

// SaveToFile saves the cache to a file
func (c *Cache) SaveToFile() error {
	jsonBytes, err := json.Marshal(c.filePathToHash)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	cacheFilePath, err := getFileCachePath()
	if err != nil {
		return err
	}
	readAndWritePermissions := os.FileMode(0644)
	if err := ioutil.WriteFile(cacheFilePath, jsonBytes, readAndWritePermissions); err != nil {
		return errors.Wrap(err, 0)
	}
	return nil
}

func getFileCachePath() (string, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	return fmt.Sprintf("%s/.s3deploy", homeDir), nil
}

func getFilePathToHashFromFile() (map[string]string, error) {
	filePathToHash := map[string]string{}
	cacheFilePath, err := getFileCachePath()
	if err != nil {
		return filePathToHash, err
	}
	var jsonBytes []byte
	jsonBytes, err = ioutil.ReadFile(cacheFilePath)
	if err != nil {
		return filePathToHash, errors.Wrap(err, 0)
	}
	json.Unmarshal(jsonBytes, &filePathToHash)
	return filePathToHash, nil
}

func readJSONFile(filePath string, holder interface{}) (err error) {
	var jsonBytes []byte
	jsonBytes, err = ioutil.ReadFile(filePath)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	json.Unmarshal(jsonBytes, holder)

	return nil
}

func createCacheFileInNonExists() error {
	cacheFilePath, err := getFileCachePath()
	if err != nil {
		return err
	}

	// detect if file exists
	_, err = os.Stat(cacheFilePath)
	// If not, create the file
	if os.IsNotExist(err) {
		var file *os.File
		file, err = os.Create(cacheFilePath)
		if err != nil {
			return errors.Wrap(err, 0)
		}
		if _, err = file.WriteString("{}"); err != nil {
			return errors.Wrap(err, 0)
		}
		defer file.Close()
	}
	return nil
}

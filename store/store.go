package store

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"

	"github.com/go-errors/errors"
)

// Store manages the cache and keeps track of which files must be uploaded to s3 and which removed
type Store struct {
	cache                *Cache
	filePathToWasTouched map[string]bool
}

// New creates a new store
func New() (*Store, error) {
	cache, err := NewCache()
	if err != nil {
		return nil, err
	}
	filePathToWasTouched := map[string]bool{}
	for _, filePath := range cache.GetFilePaths() {
		filePathToWasTouched[filePath] = false
	}

	return &Store{
		cache:                cache,
		filePathToWasTouched: filePathToWasTouched,
	}, nil
}

// DidFileChange returns whether or not the file had changed since the last time s3deploy was run
func (s *Store) DidFileChange(filePath string) (bool, error) {
	newHash, err := getFileHash(filePath)
	if err != nil {
		return false, err
	}
	oldHash, ok := s.cache.Get(filePath)
	// If file did not exist before, we consider it "changed"
	if !ok {
		return true, nil
	}
	didFileChange := newHash != oldHash
	return didFileChange, nil
}

// UpdateCache updates the hash stored in the cache for a given filePath
func (s *Store) UpdateCache(filePath string) error {
	hash, err := getFileHash(filePath)
	if err != nil {
		return err
	}
	if err := s.cache.Set(filePath, hash); err != nil {
		return err
	}
	return nil
}

// MarkAsTouched marks a file as seen during this s3deploy request
func (s *Store) MarkAsTouched(filePath string) {
	s.filePathToWasTouched[filePath] = true
}

// GetFilePathsToDeleteOnRemote returns the filePaths that were stored in the cache but no longer
// referenced in the last deploy and have therefore been deleted locally.
func (s *Store) GetFilePathsToDeleteOnRemote() (filePathsToDelete []string) {
	for filePath, wasTouched := range s.filePathToWasTouched {
		if !wasTouched {
			filePathsToDelete = append(filePathsToDelete, filePath)
		}
	}
	return filePathsToDelete
}

// SaveCacheToFile saves the state of the cache to a file
func (s *Store) SaveCacheToFile() error {
	if err := s.cache.SaveToFile(); err != nil {
		return err
	}
	return nil
}

func getFileHash(filePath string) (string, error) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", errors.Wrap(err, 0)
	}
	hashInBytes := sha256.Sum256(file)
	hash := fmt.Sprintf("%x", hashInBytes) // base-16 string w/ lowercase letters
	return hash, nil
}

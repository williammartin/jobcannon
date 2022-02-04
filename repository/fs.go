package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/williammartin/jobcannon"
)

var errorAbsent = errors.New("record is absent")

func IsAbsent(err error) bool {
	return errors.Is(err, errorAbsent)
}

type Storage struct {
	dirPath string
}

func (s *Storage) filePath(catalogId jobcannon.CatalogId, jobId jobcannon.JobId) string {
	return filepath.Join(s.dirPath, fmt.Sprintf("%d-%d.json", catalogId, jobId))
}

func Filesystem(dirPath string) (*Storage, error) {
	storageDirExists, err := dirExists(dirPath)
	if err != nil {
		return &Storage{}, fmt.Errorf("unexpected error while checking existence of %s directory: %w", dirPath, err)
	}

	if !storageDirExists {
		return &Storage{}, fmt.Errorf("storage directory '%s' does not exist", dirPath)
	}

	return &Storage{dirPath: dirPath}, nil
}

func (s *Storage) Persist(review jobcannon.ExpressionOfInterest) error {
	marshalledReview, err := json.Marshal(review)
	if err != nil {
		return fmt.Errorf("failed to marshal job review %v: %w", review, err)
	}

	if err := ioutil.WriteFile(s.filePath(review.CatalogId, review.JobId), marshalledReview, 0644); err != nil {
		return fmt.Errorf("failed to persist to filesystem: %w", err)
	}

	return nil
}

func (s *Storage) Exists(catalogId jobcannon.CatalogId, jobId jobcannon.JobId) (bool, error) {
	return fileExists(s.filePath(catalogId, jobId))
}

func (s *Storage) Load(catalogId jobcannon.CatalogId, jobId jobcannon.JobId) (jobcannon.ExpressionOfInterest, error) {
	content, err := ioutil.ReadFile(s.filePath(catalogId, jobId))
	if os.IsNotExist(err) {
		return jobcannon.ExpressionOfInterest{}, errorAbsent
	}

	if err != nil {
		return jobcannon.ExpressionOfInterest{}, fmt.Errorf("failed to load content from filesystem: %w", err)
	}

	var jobReview *jobcannon.ExpressionOfInterest = new(jobcannon.ExpressionOfInterest)
	if err := json.Unmarshal(content, jobReview); err != nil {
		return jobcannon.ExpressionOfInterest{}, fmt.Errorf("failed to load marshal file content: %w", err)
	}

	return *jobReview, nil
}

func dirExists(dirPath string) (bool, error) {
	f, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if !f.IsDir() {
		return false, fmt.Errorf("path exists but was not a directory")
	}

	return true, nil
}

func fileExists(filePath string) (bool, error) {
	f, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if f.IsDir() {
		return false, fmt.Errorf("path exists but was a directory")
	}

	return true, nil
}

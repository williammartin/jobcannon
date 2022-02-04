package repository_test

import (
	"fmt"
	"testing"

	"github.com/matryer/is"
	"github.com/williammartin/jobcannon"
	"github.com/williammartin/jobcannon/repository"
)

func TestNonExistentDir(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	_, err := repository.Filesystem("non-existent-dir")
	is.Equal(err, fmt.Errorf("storage directory 'non-existent-dir' does not exist"))
}

func TestPersistAndLoad(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	fs, err := repository.Filesystem(t.TempDir())
	is.NoErr(err)

	reviewToPersist := jobcannon.ExpressionOfInterest{
		CatalogId:  1,
		JobId:      2,
		By:         "Will",
		Text:       "Job Description Text",
		Interested: true,
	}
	err = fs.Persist(reviewToPersist)
	is.NoErr(err)

	loadedReview, err := fs.Load(reviewToPersist.CatalogId, reviewToPersist.JobId)
	is.NoErr(err)

	is.Equal(reviewToPersist, loadedReview)
}

func TestExistingFile(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	fs, err := repository.Filesystem(t.TempDir())
	is.NoErr(err)

	reviewToPersist := jobcannon.ExpressionOfInterest{
		CatalogId:  1,
		JobId:      2,
		By:         "Will",
		Text:       "Job Description Text",
		Interested: true,
	}
	err = fs.Persist(reviewToPersist)
	is.NoErr(err)

	exists, err := fs.Exists(1, 2)
	is.NoErr(err)
	is.True(exists)
}

func TestNonExistingFile(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	fs, err := repository.Filesystem(t.TempDir())
	is.NoErr(err)

	exists, err := fs.Exists(1, 2)
	is.NoErr(err)
	is.True(!exists)
}

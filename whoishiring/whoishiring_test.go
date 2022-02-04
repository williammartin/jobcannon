package whoishiring_test

import (
	"testing"

	"github.com/matryer/is"
	"github.com/williammartin/jobcannon/whoishiring"
)

// Silly integration test since the WhoIsHiring data is not really
// under our control and I'm not going to bother faking it.
// Just going to rely on the types tying things together.
func TestFetchMostRecentJobPost(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	client := whoishiring.Client{}
	catalog, err := client.FetchMostRecentCatalog()
	is.NoErr(err)
	is.True(len(catalog.JobIds) > 0)
}

func TestFetchJobPosting(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	client := whoishiring.Client{}
	catalog, err := client.FetchMostRecentCatalog()
	is.NoErr(err)
	is.True(len(catalog.JobIds) > 0)

	_, err = client.FetchJob(catalog.JobIds[0])
	is.NoErr(err)

	// Probably should test something here on the content of job posting...
}

package whoishiring

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"

	"github.com/williammartin/jobcannon"
)

const HACKERNEWS_API api_root_url = "https://hacker-news.firebaseio.com/v0"

type api_root_url string
type api_resource_url string

func (u api_root_url) user(id string) api_resource_url {
	return api_resource_url(fmt.Sprintf("%s/user/%s.json", u, id))
}

func (u api_root_url) item(id int) api_resource_url {
	return api_resource_url(fmt.Sprintf("%s/item/%d.json", u, id))
}

// Consider tying this more closely to the resource type to guarantee that we have the right shape to unmarshal into
func fetch(url api_resource_url, target interface{}) error {
	r, err := http.Get(string(url))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

const WHO_IS_HIRING_ACCOUNT_ID = "whoishiring"

// Hackernews API Types
type account struct {
	Id        string `json:"id"`
	Submitted []int  `json:"submitted"`
}

type story struct {
	Id   int   `json:"id"`
	Kids []int `json:"kids"`
}

type comment struct {
	Id   int    `json:"id"`
	By   string `json:"by"`
	Text string `json:"text"`
}

type Client struct{}

func (c *Client) FetchMostRecentCatalog() (jobcannon.Catalog, error) {
	var fetchedAccount *account = new(account)
	if err := fetch(HACKERNEWS_API.user(WHO_IS_HIRING_ACCOUNT_ID), fetchedAccount); err != nil {
		return jobcannon.Catalog{}, fmt.Errorf("failed to fetch the account details for WhoIsHiring: %w", err)
	}

	if len(fetchedAccount.Submitted) == 0 {
		return jobcannon.Catalog{}, fmt.Errorf("expected WhoIsHiring to have at least one submission: %v", fetchedAccount)
	}
	mostRecentSubmission := fetchedAccount.Submitted[0]

	var fetchedStory *story = new(story)
	if err := fetch(HACKERNEWS_API.item(mostRecentSubmission), fetchedStory); err != nil {
		return jobcannon.Catalog{}, fmt.Errorf("failed to fetch the post details for item %d: %w", mostRecentSubmission, err)
	}

	return jobcannon.Catalog{
		Id:     jobcannon.CatalogId(fetchedStory.Id),
		JobIds: arrayMap(intToJobID, fetchedStory.Kids),
	}, nil
}

func (c *Client) FetchJob(jobId jobcannon.JobId) (jobcannon.Job, error) {
	var fetchedComment *comment = new(comment)
	if err := fetch(HACKERNEWS_API.item(int(jobId)), fetchedComment); err != nil {
		return jobcannon.Job{}, fmt.Errorf("failed to fetch job comment with id %d: %w", jobId, err)
	}

	return jobcannon.Job{
		Id:   jobcannon.JobId(fetchedComment.Id),
		By:   fetchedComment.By,
		Text: html.UnescapeString(fetchedComment.Text),
	}, nil
}

func intToJobID(id int) jobcannon.JobId {
	return jobcannon.JobId(id)
}

func arrayMap[A any, B any](mapFn func(A) B, as []A) []B {
	mappedVals := []B{}
	for _, a := range as {
		mappedVals = append(mappedVals, mapFn(a))
	}
	return mappedVals
}

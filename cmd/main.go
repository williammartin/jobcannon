package main

import (
	"os"
	"path"

	"github.com/williammartin/jobcannon"
	"github.com/williammartin/jobcannon/repository"
	"github.com/williammartin/jobcannon/ui"
	"github.com/williammartin/jobcannon/whoishiring"
)

type Repository interface {
	Persist(expressionOfInterest jobcannon.ExpressionOfInterest) error
	Exists(catalogId jobcannon.CatalogId, jobId jobcannon.JobId) (bool, error)
	Load(catalogId jobcannon.CatalogId, jobId jobcannon.JobId) (jobcannon.ExpressionOfInterest, error)
}

type Source interface {
	FetchMostRecentCatalog() (jobcannon.Catalog, error)
	FetchJob(jobId jobcannon.JobId) (jobcannon.Job, error)
}

type TUI interface {
}

type UI interface {
	DisplayText(text string)
	DisplayNewLine()
	PromptForConfirmation(text string) bool
}

type FakeExpressionRecorder struct {
	calledWith jobcannon.ExpressionOfInterest
}

func (fer *FakeExpressionRecorder) RecordExpression(expressionOfInterest jobcannon.ExpressionOfInterest) error {
	fer.calledWith = expressionOfInterest
	return nil
}

func (fer *FakeExpressionRecorder) WasCalledWith(expected jobcannon.ExpressionOfInterest) bool {
	return fer.calledWith == expected
}

func main() {
	appModel := ui.NewApplicationModel(&FakeExpressionRecorder{}, 1, []jobcannon.Job{
		{
			Id:   0,
			By:   "will",
			Text: "foo",
		},
	})
	err := ui.TUIApplicationView(appModel)
	mustNot(err)

	// repo := createRepository()
	// source := createSource()
	// ui := createUI()

	// catalog, err := source.FetchMostRecentCatalog()
	// mustNot(err)

	// for _, jobId := range catalog.JobIds {
	// 	exists, err := repo.Exists(catalog.Id, jobId)
	// 	mustNot(err)

	// 	if !exists {
	// 		job, err := source.FetchJob(jobId)
	// 		mustNot(err)

	// 		ui.DisplayText(job.Text)
	// 		ui.DisplayNewLine()
	// 		ui.DisplayNewLine()
	// 		interested := ui.PromptForConfirmation("Are you interested in this job?")
	// 		repo.Persist(jobcannon.ExpressionOfInterest{
	// 			CatalogId:  catalog.Id,
	// 			JobId:      job.Id,
	// 			By:         job.By,
	// 			Text:       job.Text,
	// 			Interested: interested,
	// 		})
	// 		ui.DisplayNewLine()
	// 	}
	// }
}

// These create functions are kind of silly, I'm mainly using them to type check
// the interfaces until I extract different CLI command structures.
func createRepository() Repository {
	defaultUserConfigDir, err := os.UserConfigDir()
	mustNot(err)

	jobcannonDir := path.Join(defaultUserConfigDir, "jobcannon")

	err = os.MkdirAll(jobcannonDir, 0755)
	mustNot(err)

	fs, err := repository.Filesystem(jobcannonDir)
	mustNot(err)

	return fs
}

func createSource() Source {
	return &whoishiring.Client{}
}

func createUI() UI {
	return &ui.Console{}
}

func mustNot(err error) {
	if err != nil {
		panic(err)
	}
}

package ui_test

import (
	"testing"

	"github.com/matryer/is"
	"github.com/reactivex/rxgo/v2"
	"github.com/williammartin/jobcannon"
	"github.com/williammartin/jobcannon/ui"
)

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

// Catalog View Model
func TestObservingNoJobs(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	app := ui.NewApplicationModel(&FakeExpressionRecorder{}, 1, []jobcannon.Job{})
	jobsObs := app.CatalogModel().ViewModel().Jobs().Observe()
	jobModels := latestJobs(t, jobsObs)

	is.True(len(jobModels) == 0)
}

func TestObservingOneJob(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	app := ui.NewApplicationModel(&FakeExpressionRecorder{}, 1, []jobcannon.Job{{
		Id:   1,
		By:   "will",
		Text: "will's text",
	}})
	jobsObs := app.CatalogModel().ViewModel().Jobs().Observe()
	jobModels := latestJobs(t, jobsObs)

	is.True(len(jobModels) == 1)
	is.Equal(jobModels[0].ViewModel().By(), "will")
}

func TestExpressingInterestRemovesJobFromCatalogViewModel(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	app := ui.NewApplicationModel(&FakeExpressionRecorder{}, 1, []jobcannon.Job{{
		Id:   1,
		By:   "will",
		Text: "will's text",
	}})
	jobsObs := app.CatalogModel().ViewModel().Jobs().Observe()
	jobModels := latestJobs(t, jobsObs)
	is.True(len(jobModels) == 1)

	jobModels[0].ExpressInterest(true)
	is.True(len(latestJobs(t, jobsObs)) == 0)
}

func TestExpressingInterestRecordsExpressionWithCollaborator(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	expressionRecorderSpy := &FakeExpressionRecorder{}
	app := ui.NewApplicationModel(expressionRecorderSpy, 1, []jobcannon.Job{{
		Id:   2,
		By:   "will",
		Text: "will's text",
	}})
	jobsObs := app.CatalogModel().ViewModel().Jobs().Observe()
	jobModels := latestJobs(t, jobsObs)
	jobModels[0].ExpressInterest(true)

	is.True(expressionRecorderSpy.WasCalledWith(jobcannon.ExpressionOfInterest{
		CatalogId:  1,
		JobId:      2,
		By:         "will",
		Text:       "will's text",
		Interested: true,
	}))
}

// Consider a test harness that allows us to write more expressive assertions
// like t.expect.latestJobs(...)

func latestJobs(t *testing.T, jobsObs <-chan rxgo.Item) []ui.JobModel {
	is := is.New(t)
	is.Helper()

	observedValue := <-jobsObs
	if observedValue.Error() {
		is.Fail()
	}
	return observedValue.V.([]ui.JobModel)
}

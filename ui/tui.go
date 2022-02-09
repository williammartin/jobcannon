package ui

import (
	"context"
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/reactivex/rxgo/v2"
	"github.com/williammartin/jobcannon"
	"github.com/williammartin/jobcannon/slice"
)

type ExpressionRecorder interface {
	RecordExpression(expressionOfInterest jobcannon.ExpressionOfInterest) error
}

// A lot of types in here that look dumb, and perhaps they are dumb -
// particularly the models having almost no behaviour except providing view models.
// On the other hand, perhaps if the application were to become more complicated,
// having this split already would be handy.

// Not sure any of these interfaces are really necessary...we can maybe just use concrete types in the view side...
type CatalogModel interface {
	ViewModel() CatalogViewModel
}

type ICatalogModel struct {
	jobModelsObs rxgo.Observable

	expressionRecorder ExpressionRecorder
}

func (cm *ICatalogModel) ViewModel() CatalogViewModel {
	return &ICatalogViewModel{
		jobs: cm.jobModelsObs,
	}
}

type CatalogViewModel interface {
	Jobs() rxgo.Observable
}

type ICatalogViewModel struct {
	jobs rxgo.Observable
}

func (cvm *ICatalogViewModel) Jobs() rxgo.Observable {
	return cvm.jobs
}

type JobModel interface {
	ExpressInterest(interested bool)
	ViewModel() JobViewModel
}

type IJobModel struct {
	job             jobcannon.Job
	expressInterest func(interested bool)
}

func (jm *IJobModel) ExpressInterest(interested bool) {
	jm.expressInterest(interested)
}

func (jm *IJobModel) ViewModel() JobViewModel {
	return &IJobViewModel{
		by:   jm.job.By,
		text: jm.job.Text,
	}
}

type JobViewModel interface {
	By() string
	Text() string
	// Targeted() ?
	// Handled() ?
}

type IJobViewModel struct {
	by   string
	text string
}

func (jvm *IJobViewModel) By() string {
	return jvm.by
}

func (jvm *IJobViewModel) Text() string {
	return jvm.text
}

type ApplicationModel struct {
	jobsCh             chan rxgo.Item
	catalogModel       CatalogModel
	expressionRecorder ExpressionRecorder
}

func (am *ApplicationModel) LoadJobs() {
	am.jobsCh <- rxgo.Of([]jobcannon.Job{
		{
			Id:   0,
			By:   "cole",
			Text: "foo",
		},
		{
			Id:   1,
			By:   "will",
			Text: "foo",
		},
		{
			Id:   2,
			By:   "dylan",
			Text: "foo",
		},
	})
}

// This is very much a first stab at using MVVM and Reactive Programming in Go.
// It seems pretty gross with the lack of generics (unsafe type assertions), and unpleasant syntax for anonymous functions
func NewApplicationModel(expressionRecorder ExpressionRecorder, catalogId jobcannon.CatalogId, jobs []jobcannon.Job) *ApplicationModel {
	// Gross mutability, but I don't know if it's possible to get the current value from the observable...
	currentJobs := jobs
	jobsCh := make(chan rxgo.Item)
	catalogModel := &ICatalogModel{
		jobModelsObs: rxgo.FromChannel(jobsCh).Map(func(_ context.Context, item interface{}) (interface{}, error) {
			jobs := item.([]jobcannon.Job)
			return slice.Map(func(job jobcannon.Job) JobModel {
				return &IJobModel{
					job: job,
					expressInterest: func(interested bool) {
						expressionRecorder.RecordExpression(jobcannon.ExpressionOfInterest{
							CatalogId:  catalogId,
							JobId:      job.Id,
							By:         job.By,
							Text:       job.Text,
							Interested: interested,
						})

						currentJobs = slice.Drop(func(j jobcannon.Job) bool {
							return job.Id == j.Id
						}, currentJobs)
						jobsCh <- rxgo.Of(currentJobs)
					},
				}
			}, jobs), nil
		}),
		expressionRecorder: expressionRecorder,
	}

	return &ApplicationModel{
		jobsCh:       jobsCh,
		catalogModel: catalogModel,
	}
}

func (am *ApplicationModel) CatalogModel() CatalogModel {
	return am.catalogModel
}

// TUI View

func TUIApplicationView(applicationModel *ApplicationModel) error {
	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		return fmt.Errorf("could not create GUI: %w", err)
	}
	defer g.Close()
	g.Cursor = true
	g.Highlight = true
	g.BgColor = gocui.ColorBlack
	g.SelBgColor = gocui.ColorGreen
	g.SelFgColor = gocui.ColorBlack

	maxX, maxY := g.Size()
	appView, err := g.SetView("application", -1, -1, maxX, maxY)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	if err := TUICatalogView(g, appView, applicationModel.CatalogModel().ViewModel()); err != nil {
		return err
	}

	applicationModel.LoadJobs()

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		return err
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}

	return nil
}

func TUICatalogView(g *gocui.Gui, appView *gocui.View, catalogViewModel CatalogViewModel) error {
	maxX, maxY := appView.Size()
	debugView, err := g.SetView("debug", -1, maxY-5, maxX, maxY)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	backlogView, err := g.SetView("backlog", -1, 1, int(0.1*float32(maxX)), maxY-5)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	backlogView.Title = "Jobs To Review"

	detailsView, err := g.SetView("details", int(0.1*float32(maxX)), 1, maxX, maxY-5)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	detailsView.Title = "Job Details"

	if err := g.SetKeybinding("backlog", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}

	if err := g.SetKeybinding("backlog", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}

	if _, err := g.SetCurrentView("backlog"); err != nil {
		return err
	}

	// Sus that this might exit with a buffered channel...
	catalogViewModel.Jobs().DoOnNext(func(observedValue interface{}) {
		fmt.Fprintf(debugView, "received")
		jobs := observedValue.([]JobModel)
		for _, job := range jobs {
			fmt.Fprintln(backlogView, job.ViewModel().By())
		}
	})

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

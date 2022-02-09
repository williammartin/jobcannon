package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/jroimartin/gocui"
	"github.com/williammartin/jobcannon"
	"github.com/williammartin/jobcannon/repository"
	"github.com/williammartin/jobcannon/slice"
	"github.com/williammartin/jobcannon/ui"
	"github.com/williammartin/jobcannon/whoishiring"
	"jaytaylor.com/html2text"
)

type Msg interface {
	sealed()
}

type CursorDown struct{}

func (*CursorDown) sealed() {}

type CursorUp struct{}

func (*CursorUp) sealed() {}

type AppModel struct {
	items []string
}

func Update()

func main() {
	repo := createRepository()
	source := createSource()

	catalog, err := source.FetchMostRecentCatalog()
	mustNot(err)

	jobs := []jobcannon.Job{}
	for _, jobId := range catalog.JobIds[0:1] {
		exists, err := repo.Exists(catalog.Id, jobId)
		mustNot(err)

		if !exists {
			job, err := source.FetchJob(jobId)
			mustNot(err)
			jobs = append(jobs, job)
			// repo.Persist(jobcannon.ExpressionOfInterest{
			// 	CatalogId:  catalog.Id,
			// 	JobId:      job.Id,
			// 	By:         job.By,
			// 	Text:       job.Text,
			// 	Interested: interested,
			// })
		}
	}

	// GUI Stuff
	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true
	g.Highlight = true
	g.BgColor = gocui.ColorBlack
	g.SelBgColor = gocui.ColorGreen
	g.SelFgColor = gocui.ColorBlack

	// TODO: Figure out some reasonable sizes that aren't magic numbers
	maxX, maxY := g.Size()

	reviewedView, err := g.SetView("backlog", 0, 0, int(0.1*float32(maxX)), maxY-21)
	if err != nil && err != gocui.ErrUnknownView {
		panic(err)
	}
	reviewedView.Title = "Backlog"

	detailsView, err := g.SetView("details", int(0.1*float32(maxX)+2), 0, maxX-1, maxY-21)
	if err != nil && err != gocui.ErrUnknownView {
		panic(err)
	}
	detailsView.Title = "Job Details"
	detailsView.Wrap = true

	debugView, err := g.SetView("debug", 0, maxY-20, maxX-1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		panic(err)
	}
	debugView.Title = "Debugger"
	debugView.Autoscroll = true

	ui.ListView(g, ui.AbsolutePosition{
		X0: 0,
		Y0: 0,
		X1: int(0.1 * float32(maxX)),
		Y1: maxY - 22,
	}, ui.NewListViewModel("Jobs", slice.Map(func(job jobcannon.Job) ui.ListItemViewModel {
		label := fmt.Sprintf("%d - %s", job.Id, job.By)
		return ui.ListItemViewModel{
			Label: label,
			OnSelectedFn: func() {
				fmt.Fprintf(debugView, "Selected %s\n", label)
				g.Update(func(g *gocui.Gui) error {
					v, err := g.View("details")
					if err != nil {
						// log to debugger
						return err
					}
					v.Clear()
					text, err := html2text.FromString(job.Text, html2text.Options{PrettyTables: true})
					if err != nil {
						// log to debugger
						return err
					}

					fmt.Fprintln(v, text)
					return nil
				})
			},
			OnAcceptedFn: func() {
				fmt.Fprintf(debugView, "Accepted %s\n", label)
			},
			OnRemovedFn: func() {
				fmt.Fprintf(debugView, "Removed %s\n", label)
			},
		}
	}, jobs)))

	if _, err := g.SetCurrentView("Jobs"); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		panic(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(err)
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

type Repository interface {
	Persist(expressionOfInterest jobcannon.ExpressionOfInterest) error
	Exists(catalogId jobcannon.CatalogId, jobId jobcannon.JobId) (bool, error)
	Load(catalogId jobcannon.CatalogId, jobId jobcannon.JobId) (jobcannon.ExpressionOfInterest, error)
}

type Source interface {
	FetchMostRecentCatalog() (jobcannon.Catalog, error)
	FetchJob(jobId jobcannon.JobId) (jobcannon.Job, error)
}

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

func mustNot(err error) {
	if err != nil {
		panic(err)
	}
}

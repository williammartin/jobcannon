package ui

import (
	"github.com/jroimartin/gocui"
)

type PanelViewModel struct {
	Name  string
	Title string
	// Make text observable later...
	Text string
}

func NewPanelViewModel(name string, title string, text string) *PanelViewModel {
	return &PanelViewModel{
		Title: title,
		Text:  text,
	}
}

func PanelView(g *gocui.Gui, pos AbsolutePosition, panelVM *PanelViewModel) error {
	v, err := g.SetView(panelVM.Name, pos.X0, pos.Y0, pos.X1, pos.Y1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	v.Frame = true

	return nil
}

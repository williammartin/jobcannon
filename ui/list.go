package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/reactivex/rxgo/v2"
)

type ListItemViewModel struct {
	Label        string
	OnSelectedFn func()
	OnAcceptedFn func()
	OnRemovedFn  func()
}

// If we make the Items observable and add a "selected" boolean,
// we could render a Y/N next to the label.
type ListViewModel struct {
	Name                     string
	Items                    []ListItemViewModel
	CurrentSelectionIndexObs rxgo.Observable

	// Private
	// If RxGo doesn't have a way to observe a behavior like RxJS, we're
	// going to have to carry things around and mutate them, and then put
	// them on a channel which feels very redundant.
	currentSelectionIndex   int
	currentSelectionIndexCh chan rxgo.Item
}

func NewListViewModel(name string, items []rxgo.Observable) *ListViewModel {
	currentSelectionIndex := 0
	currentSelectionIndexCh := make(chan rxgo.Item)

	go func() {
		currentSelectionIndexCh <- rxgo.Of(currentSelectionIndex)
		items[currentSelectionIndex].OnSelectedFn()
	}()

	// go func() {
	// Need to subscribe to the currentSelectionIndex and then call
	// vm.Items[vm.currentSelectionIndex].OnSelectedFn()
	// For the moment, we'll do it manually above
	// }()

	return &ListViewModel{
		Name:                     name,
		Items:                    items,
		CurrentSelectionIndexObs: rxgo.FromChannel(currentSelectionIndexCh),

		currentSelectionIndex:   currentSelectionIndex,
		currentSelectionIndexCh: currentSelectionIndexCh,
	}
}

func (vm *ListViewModel) Next() {
	vm.CurrentSelectionIndexObs.Last()
	// This list will not wrap
	if vm.currentSelectionIndex == len(vm.Items)-1 {
		return
	}

	vm.currentSelectionIndex = vm.currentSelectionIndex + 1
	vm.currentSelectionIndexCh <- rxgo.Of(vm.currentSelectionIndex)
	vm.Items[vm.currentSelectionIndex].OnSelectedFn()
}

func (vm *ListViewModel) Previous() {
	// This list will not wrap
	if vm.currentSelectionIndex == 0 {
		return
	}

	vm.currentSelectionIndex = vm.currentSelectionIndex - 1
	vm.currentSelectionIndexCh <- rxgo.Of(vm.currentSelectionIndex)
	vm.Items[vm.currentSelectionIndex].OnSelectedFn()
}

func (vm *ListViewModel) AcceptCurrent() {
	vm.Items[vm.currentSelectionIndex].OnAcceptedFn()
}

func (vm *ListViewModel) RemoveCurrent() {
	vm.Items[vm.currentSelectionIndex].OnRemovedFn()
}

func ListView(g *gocui.Gui, pos AbsolutePosition, listVM *ListViewModel) error {
	v, err := g.SetView(listVM.Name, pos.X0, pos.Y0, pos.X1, pos.Y1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	v.Frame = false

	// Here lies the behaviour for the view.
	// It sends commands to the ViewModel, and then updates the view as a result of changes to the state of the ViewModel.
	if err := g.SetKeybinding(listVM.Name, gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		// THINK: Do we need to check whether v is nil here?
		// THINK: Do we need to SetOrigin?
		// Sort of doing BLoC pattern here, using the CurrentItemIndex to position the cursor, but it's Mutating rather than
		// nicely reactive. Consider whether an Observable here would be nice?
		listVM.Next()
		return nil
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding(listVM.Name, gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		listVM.Previous()
		return nil
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding(listVM.Name, gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		listVM.AcceptCurrent()
		return nil
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding(listVM.Name, gocui.KeyBackspace2, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		listVM.RemoveCurrent()
		return nil
	}); err != nil {
		return err
	}

	// Now let's OBSERVE BOYYYY
	listVM.CurrentSelectionIndexObs.DoOnNext(func(observedValue interface{}) {
		g.Update(func(g *gocui.Gui) error {
			v, err := g.View(listVM.Name)
			if err != nil {
				// log to debugger
				return err
			}
			currentIndex := observedValue.(int)
			return v.SetCursor(pos.X0, pos.Y0+currentIndex)
		})
	})

	// Print each Item in the list
	for _, item := range listVM.Items {
		fmt.Fprintln(v, item.Label)
	}

	return nil
}

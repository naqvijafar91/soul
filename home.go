package soul

import (
	"fmt"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Home struct {
	Text    *Note
	Service *NoteService

	selectedNote *Note
	textWidget   *widget.Entry
	infoLabel    *widget.Label
	listWidget   *widget.List
}

const DefaultInfo = "Welcome to your soul"

func (home *Home) addNote() error {
	newNote, err := home.Service.Create()
	if err != nil {
		return err
	}

	home.setNoteAndBind(newNote)

	return nil
}

func (ui *Home) setNoteAndBind(n *Note) {
	ui.textWidget.Unbind()
	if n == nil {
		ui.textWidget.SetText(ui.placeholderContent())
		return
	}
	ui.Text = n
	ui.textWidget.Bind(n.Text)
	ui.textWidget.Validator = nil
	ui.listWidget.Refresh()
	ui.selectedNote = n
}

func (ui *Home) buildList(notes []Note) *widget.List {
	list := widget.NewList(
		func() int {
			return len(ui.Service.Notes)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Title")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			note := ui.Service.Notes[id]
			label.Bind(note.title())
		})

	list.OnSelected = func(id widget.ListItemID) {
		n := ui.Service.Notes[id]
		ui.setNoteAndBind(&n)
	}

	return list
}

func (ui *Home) LoadDataAndBuildUI() (fyne.CanvasObject, error) {
	ui.textWidget = widget.NewMultiLineEntry()
	ui.textWidget.Wrapping = fyne.TextWrapWord
	ui.textWidget.SetText(ui.placeholderContent())
	ui.infoLabel = widget.NewLabel("Welcome to your Soul")

	err := ui.Service.LoadAll()
	if err != nil {
		return nil, err
	}

	ui.listWidget = ui.buildList(ui.Service.Notes)
	notes := ui.Service.Notes
	if len(notes) > 0 {
		ui.setNoteAndBind(&ui.Service.Notes[0])
		ui.listWidget.Select(0)
	}

	var updateNote = func(note *Note, disableDuringOp bool) {
		ui.infoLabel.SetText("updating note.....")
		defer func() {
			ui.infoLabel.SetText(DefaultInfo)
			
			if ui.textWidget.Disabled() {
				ui.textWidget.Enable()
			}
		}()

		if disableDuringOp {
			ui.textWidget.Disable()
		}

		err := ui.Service.Update(note)
		if err != nil {
			fyne.CurrentApp().SendNotification(fyne.NewNotification("Unable to save this note", fmt.Sprintf("%v", err)))
		}
	}

	bar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			err := ui.addNote()
			if err != nil {
				ui.infoLabel.SetText(fmt.Sprintf("failed to create new note %v", err))
			}
		}),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			updateNote(ui.selectedNote, true)
		}),
	)

	side := container.New(layout.NewBorderLayout(bar, ui.infoLabel, nil, nil),
		bar, container.NewVScroll(ui.listWidget), ui.infoLabel)

	// finally start our sync service
	ss := NewSyncService(func() ([]Note, error) {
		return ui.Service.Notes, nil
	}, func(note *Note) error {
		updateNote(note, false)
		return nil
	}, func(err error) {
		ui.infoLabel.SetText(fmt.Sprintf("failed to sync note %v", err))
	}, 5*time.Second)
	ss.Start()

	// TODO : fix correct size and disable resie window
	return newAdaptiveSplit(side, ui.textWidget), nil
}

func (ui *Home) RegisterKeys(w fyne.Window) {
	shortcut := &desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: fyne.KeyModifierControl}
	if runtime.GOOS == "darwin" {
		shortcut.Modifier = fyne.KeyModifierSuper
	}

	w.Canvas().AddShortcut(shortcut, func(_ fyne.Shortcut) {
		ui.addNote()
	})
}

func (ui *Home) placeholderContent() string {
	text := "Welcome!\nTap '+' in the toolbar to add a note."
	if fyne.CurrentDevice().HasKeyboard() {
		modifier := "ctrl"
		if runtime.GOOS == "darwin" {
			modifier = "cmd"
		}
		text += fmt.Sprintf("\n\nOr use the keyboard shortcut %s+N.", modifier)
	}
	return text
}

package fyne

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func NewLoginPage(currentDbPath string,
	onSubmitFunc func(email, password, updatedDbPath string, stayLoggedIn bool) error) fyne.CanvasObject {
	folderIdentifier := widget.NewEntry()
	folderIdentifier.SetPlaceHolder("Folder Name")

	passwordWidget := widget.NewPasswordEntry()
	passwordWidget.SetPlaceHolder("Password")

	dbPath := widget.NewEntry()
	dbPath.SetPlaceHolder("Db File Path")
	dbPath.SetText(currentDbPath)

	stayLoggedIn := false
	savePreferencesCheckbox := widget.NewCheck("Keep It Open", func(b bool) {
		stayLoggedIn = b
	})

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Folder Name", Widget: folderIdentifier},
			{Text: "Password", Widget: passwordWidget, HintText: "Do not forget this or all data is lost"},
			{Text: "Db Path", Widget: dbPath, HintText: "Absoulte path to the db file"},
			{Widget: savePreferencesCheckbox, HintText: "Stay logged in"},
		},
	}

	submitButton := widget.NewButton("Submit", nil)
	submitButton.OnTapped = func() {
		form.Disable()
		submitButton.SetText("Logging in....")
		defer func() {
			submitButton.SetText("Submit")
			form.Enable()
		}()

		err := onSubmitFunc(folderIdentifier.Text, passwordWidget.Text, dbPath.Text, stayLoggedIn)
		if err != nil {
			fyne.CurrentApp().SendNotification(fyne.NewNotification("Unable to login",
				fmt.Sprintf("unable to login/register due to %v", err)))
		}
	}

	return container.NewVBox(form, container.NewCenter(submitButton))
}

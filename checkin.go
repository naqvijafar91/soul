package soul

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func NewCheckInPage(_ fyne.Window,
	onSubmitFunc func(password string, loginInstead bool) error) fyne.CanvasObject {

	passwordWidget := widget.NewPasswordEntry()
	passwordWidget.SetPlaceHolder("Logged In Folder Password")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Password", Widget: passwordWidget},
		},
	}

	loginPageButton := widget.NewButton("Login Instead", nil)
	loginPageButton.OnTapped = func() {
		onSubmitFunc("", true)
	}

	submitButton := widget.NewButton("Submit", nil)
	submitButton.OnTapped = func() {
		form.Disable()
		loginPageButton.Disable()
		submitButton.SetText("Logging in....")
		defer func() {
			submitButton.SetText("Submit")
			form.Enable()
			loginPageButton.Enable()
		}()

		err := onSubmitFunc(passwordWidget.Text, false)
		if err != nil {
			fyne.CurrentApp().SendNotification(fyne.NewNotification("Unable to check in",
				fmt.Sprintf("due to %v", err)))
		}
	}

	return container.NewVBox(form, container.NewCenter(
		container.NewHBox(submitButton, loginPageButton),
	))
}

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"soul"
	"soul/crypt"
	"soul/disk"
	"strings"
	"syscall"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

const DefaultBaseScopeV1 = "soul-db-draft-v1/"

func showHomePage(window fyne.Window, service *soul.NoteService, loggedOutFunc func()) error {
	notesUI := &soul.Home{Service: service, OnLoggedOut: loggedOutFunc}
	canvas, err := notesUI.LoadDataAndBuildUI()
	if err != nil {
		return err
	}

	window.SetContent(canvas)
	notesUI.RegisterKeys(window)

	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	signal.Notify(gracefulStop, syscall.SIGQUIT)
	go func() {
		<-gracefulStop
		notesUI.Logout()
		os.Exit(0)
	}()

	return nil
}

func showLoginPage(window fyne.Window, currentDbPath string, onSubmitFunc func(email, password, updatedDbPath string, stayLoggedIn bool) error) {
	canvasObj := soul.NewLoginPage(currentDbPath, onSubmitFunc)
	window.SetContent(canvasObj)
}

func showCheckInPage(window fyne.Window, onSubmitFunc func(password string, loginInstead bool) error) {
	canvasObj := soul.NewCheckInPage(window, onSubmitFunc)
	window.SetContent(canvasObj)
}

func setupDiskRepo(folderName, password, dbPath string) (soul.NoteRepository, error) {
	repo, err := disk.NewNoteRepository(dbPath, folderName, strings.TrimSpace(password), crypt.NewSoulEncrypter, crypt.NewSoulDecrypter)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// TODO: Fix re-logging bug
func main() {
	log.SetFlags(0)
	log.SetOutput(ioutil.Discard)
	log.Println("Starting soul...")
	app := fyneapp.NewWithID("org.standard.soul.app")
	app.Settings().SetTheme(theme.DarkTheme())

	window := app.NewWindow("Soul")
	window.CenterOnScreen()

	var onLoggedInFunc = func(logoutChan chan bool) func(folderName, password, updatedDbPath string, stayLoggedIn bool) error {
		return func(folderName, password, updatedDbPath string, stayLoggedIn bool) error {
			soul.StoreDbPath(app, updatedDbPath)

			repo, err := setupDiskRepo(folderName, password, updatedDbPath)
			if err != nil {
				return err
			}

			if stayLoggedIn {
				cryptor, err := crypt.NewCryptor(strings.TrimSpace(password))
				if err != nil {
					return fmt.Errorf("failed to create cryptor %w", err)
				}

				err = soul.SetCredentials(app, cryptor, &soul.Credentials{
					Identifier: folderName,
					Password:   password,
				})
				if err != nil {
					return fmt.Errorf("failed to store credentials %w", err)
				}
			}

			err = showHomePage(window, &soul.NoteService{
				Repo: repo,
			}, func() {
				logoutChan <- true
			})
			if err != nil {
				return fmt.Errorf("failed to load home page ui %v", err)
			}

			return nil
		}
	}

	logoutChan := make(chan bool)
	go func() {
		for {
			<-logoutChan
			showLoginPage(window, soul.GetDBPath(app), onLoggedInFunc(logoutChan))
		}
	}()

	if soul.IsSignedIn(app) {
		showCheckInPage(window, func(password string, loginInstead bool) error {
			if loginInstead {
				showLoginPage(window, soul.GetDBPath(app), onLoggedInFunc(logoutChan))
				return nil
			}

			cryptor, err := crypt.NewCryptor(strings.TrimSpace(password))
			if err != nil {
				return fmt.Errorf("failed to create cryptor %w", err)
			}

			credentials, err := soul.GetCredentials(app, cryptor)
			if err != nil {
				return fmt.Errorf("failed to extract credentials %w. You may try logging in instead", err)
			}

			// login use these credentials now
			repo, err := setupDiskRepo(credentials.Identifier, credentials.Password, soul.GetDBPath(app))
			if err != nil {
				return err
			}

			err = showHomePage(window, &soul.NoteService{Repo: repo}, func() {
				showLoginPage(window, soul.GetDBPath(app), onLoggedInFunc(logoutChan))
			})
			if err != nil {
				return fmt.Errorf("failed to load home page ui %v", err)
			}

			return nil
		})
	} else {
		showLoginPage(window, soul.GetDBPath(app), onLoggedInFunc(logoutChan))
	}

	window.Resize(fyne.NewSize(1000, 600))
	window.ShowAndRun()
}

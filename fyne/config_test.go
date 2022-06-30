package fyne_test

import (
	"soul/fyne"
	"soul/testhelpers"
	"testing"

	fyneapp "fyne.io/fyne/v2/app"
)

func TestFyneConfig(t *testing.T) {
	t.Parallel()

	app := fyneapp.NewWithID("org.testing.soul")

	testhelpers.ExecuteConfigStoreTests(t, fyne.NewFyneConfigStore(app))
}

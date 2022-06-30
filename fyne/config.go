package fyne

import (
	"soul"

	"fyne.io/fyne/v2"
)

var _ soul.ConfigStore = &PrefBackedConfig{}

type PrefBackedConfig struct {
	app fyne.App
}

func (pbc *PrefBackedConfig) GetString(key string) string {
	return pbc.app.Preferences().String(key)
}

func (pbc *PrefBackedConfig) SetString(key, val string) {
	pbc.app.Preferences().SetString(key, val)
}

func (pbc *PrefBackedConfig) Delete(key string) {
	pbc.app.Preferences().RemoveValue(key)
}

// NewFyneConfigStore creates a new fyne backed config store
func NewFyneConfigStore(app fyne.App) *PrefBackedConfig {
	return &PrefBackedConfig{app: app}
}

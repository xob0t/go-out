package main

import backend "github.com/xob0t/go-out-backend"

type SettingsService struct{}

func (g *SettingsService) Update(settigns backend.MergeSettings) {
	backend.GlobalSettings.MergeSettings = settigns
	backend.SaveGlobalConfig()
}

func (g *SettingsService) Get() backend.MergeSettings {
	return backend.GlobalSettings.MergeSettings
}

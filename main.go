package main

import (
	"embed"
	"fmt"
	"log"

	backend "github.com/xob0t/go-out-backend"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed build/info.json
var WailsInfoJSON string

//go:embed frontend/dist
var assets embed.FS
var title = fmt.Sprintf("go-out v" + backend.GetAppVersion(WailsInfoJSON))

// main function serves as the application's entry point. It initializes the application, creates a window,
// It subsequently runs the application and logs any error that might occur.
func main() {

	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.
	app := application.New(application.Options{
		Name:        "go-out",
		Description: "Merge Google Photos json metadata into media files",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Services: []application.Service{
			application.NewService(&SettingsService{}),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	window := app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Title:  title,
		Width:  1280,
		Height: 720,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundType:    application.BackgroundTypeTransparent,
		BackgroundColour:  application.NewRGB(27, 38, 54),
		URL:               "/",
		EnableDragAndDrop: true,
	})

	window.RegisterHook(events.Common.WindowRuntimeReady, func(e *application.WindowEvent) {
		app.Logger.Info("WindowRuntimeReady")
		backend.RestoreSettings(app, window)
	})

	app.Events.On("mergeSettingsChanged", func(e *application.WailsEvent) {
		app.Logger.Info("mergeSettingsChanged")
		app.Logger.Info("[Go] WailsEvent received", "name", e.Name, "data", e.Data, "sender", e.Sender, "cancelled", e.Cancelled)
	})

	window.On(events.Common.WindowDidMove, func(event *application.WindowEvent) {
		backend.GlobalSettings.Window.IsMaximised = window.IsMaximised()
		backend.GlobalSettings.Window.SizeW, backend.GlobalSettings.Window.SizeH = window.Size()
		backend.GlobalSettings.Window.PosX, backend.GlobalSettings.Window.PosY = window.Position()
		backend.GlobalSettings.Window.Saved = true
		// app.Logger.Info("window resized!")
		backend.SaveGlobalConfig()
	})

	window.On(events.Common.WindowFilesDropped, func(event *application.WindowEvent) {
		files := event.Context().DroppedFiles()
		app.Events.Emit(&application.WailsEvent{
			Name: "files",
			Data: files,
		})
		app.Logger.Info("Files Dropped!", "files", files)
		jsonFiles, err := backend.GetAllJsonFiles(files)
		if err != nil {
			errMsg := "Failed to scan input path(s)"
			app.Logger.Warn(errMsg)
			app.Events.Emit(&application.WailsEvent{
				Name: "log",
				Data: map[string]string{
					"level":   "error",
					"message": errMsg,
				},
			})
			return
		}
		if len(jsonFiles) == 0 {
			app.Logger.Info("No JSONs found")
			app.Events.Emit(&application.WailsEvent{
				Name: "log",
				Data: map[string]string{
					"level":   "INFO",
					"message": "No JSONs found",
				},
			})
			return
		}
		app.Logger.Info("JSONs found!", "jsons", jsonFiles)
		backend.UpdateMetadata(app, jsonFiles, backend.GlobalSettings.MergeSettings)
		logMsg := "The process is complete, click this log to expand it"
		app.Events.Emit(&application.WailsEvent{
			Name: "log",
			Data: map[string]string{
				"level":   "INFO",
				"message": logMsg,
			},
		})

	})

	// Run the application. This blocks until the application has been exited.
	err := app.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}

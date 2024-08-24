package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"time"

	backend "github.com/xob0t/go-out-backend"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

var stringEdited = "-edited"
var processEdited = true

//go:embed frontend/dist
var assets embed.FS
var title = fmt.Sprintf("go-out v" + GetAppVersion())

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

		userSettingsDirExists := Exists(UserSettingsDir)
		if !userSettingsDirExists {
			app.Logger.Info("Created a new user settings dir")
			os.MkdirAll(UserSettingsDir, os.ModePerm)
		}

		configExists := Exists(GlobalSettingsPath)
		if !configExists {
			app.Logger.Info("Created a new user settings config")
			MakeNewDefaultConfig()
		}
		file, _ := os.ReadFile(GlobalSettingsPath)
		if len(file) == 0 {
			app.Logger.Info("config file is empty")
			MakeNewDefaultConfig()
		}

		GlobalSettingsConfig, _ = ParseGlobalConfig()

		if GlobalSettingsConfig.Window.IsMaximised {
			window.Maximise()
			return
		}
		if GlobalSettingsConfig.Window.Saved {
			window.SetSize(GlobalSettingsConfig.Window.SizeW, GlobalSettingsConfig.Window.SizeH)
			window.SetPosition(GlobalSettingsConfig.Window.PosX, GlobalSettingsConfig.Window.PosY)
		}
	})

	window.On(events.Common.WindowDidMove, func(event *application.WindowEvent) {
		GlobalSettingsConfig.Window.IsMaximised = window.IsMaximised()
		GlobalSettingsConfig.Window.SizeW, GlobalSettingsConfig.Window.SizeH = window.Size()
		GlobalSettingsConfig.Window.PosX, GlobalSettingsConfig.Window.PosY = window.Position()
		GlobalSettingsConfig.Window.Saved = true
		app.Logger.Info("window resized!")
		SaveGlobalConfig()
	})

	window.On(events.Common.WindowFilesDropped, func(event *application.WindowEvent) {
		files := event.Context().DroppedFiles()
		app.Events.Emit(&application.WailsEvent{
			Name: "files",
			Data: files,
		})
		app.Logger.Info("Path Dropped!", "files", files)
		file, _ := os.Open(files[0])
		fileInfo, _ := file.Stat()
		if fileInfo.IsDir() {
			files, _ := backend.GetAllFiles(files[0])
			app.Logger.Info("Path is a dir!", "files", files)
			backend.UpdateGeoData(app, files, stringEdited, processEdited)
		} else {
			app.Logger.Warn("Path is a file! Has to be a dir!")

		}
	})

	// Create a goroutine that emits an event containing the current time every second.
	// The frontend can listen to this event and update the UI accordingly.
	go func() {
		for {
			now := time.Now().Format(time.RFC1123)
			app.Events.Emit(&application.WailsEvent{
				Name: "time",
				Data: now,
			})
			time.Sleep(time.Second)
		}
	}()

	// Run the application. This blocks until the application has been exited.
	err := app.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}

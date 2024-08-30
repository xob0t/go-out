package backend

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/tidwall/gjson"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/wailsapp/wails/v3/pkg/application"
)

var GlobalSettings GlobalSettingsT
var ConfigDir string = filepath.Join(GetUserDir(), "/.config/go-out")
var ConfigPath string = filepath.Join(ConfigDir, "config.yaml")

func ExiftoolCheck(app *application.App) bool {
	_, err := exec.LookPath("exiftool")
	if err != nil {
		app.Logger.Warn("Exiftool not found")
		app.Events.Emit(&application.WailsEvent{
			Name: "exiftoolStatus",
			Data: false,
		})
		return false
	}
	app.Logger.Info("Exiftool found")
	return true
}

func RestoreSettings(app *application.App, window *application.WebviewWindow) {
	ConfigDirExists := Exists(ConfigDir)
	if !ConfigDirExists {
		app.Logger.Info("Created a new user config dir")
		os.MkdirAll(ConfigDir, os.ModePerm)
	}
	configExists := Exists(ConfigPath)
	file, _ := os.ReadFile(ConfigPath)
	if len(file) == 0 || !configExists {
		app.Logger.Info("config not found")
		MakeNewDefaultConfig()
	}
	GlobalSettings, _ = ParseGlobalConfig()
	if GlobalSettings.Window.IsMaximised {
		window.Maximise()
		return
	}
	if GlobalSettings.Window.Saved {
		window.SetSize(GlobalSettings.Window.SizeW, GlobalSettings.Window.SizeH)
		window.SetPosition(GlobalSettings.Window.PosX, GlobalSettings.Window.PosY)
	}
}

func GetAppVersion(WailsInfoJSON string) string {
	return gjson.Get(WailsInfoJSON, "info.0000.ProductVersion").String()
}

func GetUserDir() string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return dirname
}

func ParseGlobalConfig() (GlobalSettingsT, error) {
	var c GlobalSettingsT
	var k = koanf.New(".")
	if err := k.Load(file.Provider(ConfigPath), yaml.Parser()); err != nil {
		log.Printf("error loading global config: %v", err)
		return GlobalSettingsT{}, err
	}
	err := k.Unmarshal("", &c)
	if err != nil {
		log.Printf("error Unmarshaling global config: %v", err)
		return GlobalSettingsT{}, err
	}

	return c, nil
}

func MakeNewDefaultConfig() error {
	GlobalSettings = GlobalSettingsT{}
	GlobalSettings.MergeSettings.EditedSuffix = "edited"
	GlobalSettings.MergeSettings.InferTimezoneFromGPS = true
	GlobalSettings.MergeSettings.TimezoneOffset = time.Now().Format("-0700")
	GlobalSettings.MergeSettings.OverwriteExistingTags = true
	GlobalSettings.MergeSettings.ExifTags.Title = true
	GlobalSettings.MergeSettings.ExifTags.Description = true
	GlobalSettings.MergeSettings.ExifTags.DateTaken = true
	GlobalSettings.MergeSettings.ExifTags.URL = true
	GlobalSettings.MergeSettings.ExifTags.GPS = true
	err := SaveGlobalConfig()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func SaveGlobalConfig() error {
	k := koanf.New(".")

	err := k.Load(structs.Provider(GlobalSettings, "koanf"), nil)
	if err != nil {
		fmt.Println(err)
		return err
	}
	b, err := k.Marshal(yaml.Parser())
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = os.WriteFile(ConfigPath, b, os.ModePerm)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

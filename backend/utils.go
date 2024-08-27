package backend

import (
	_ "embed"
	"fmt"
	"log"
	"os"
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

type MergeSettings struct {
	IgnoreMinorErrors    bool   `json:"ignoreMinorErrors" koanf:"ignore_minor_errors"`
	EditedSuffix         string `json:"editedSuffix" koanf:"edited_suffix"`
	TimezoneOffset       string `json:"timezoneOffset" koanf:"timezone_offset"`
	InferTimezoneFromGPS bool   `json:"inferTimezoneFromGPS" koanf:"defer_timezone_from_GPS"`
	MergeTitle           bool   `json:"mergeTitle" koanf:"merge_title"`
	MergeDescription     bool   `json:"mergeDescription" koanf:"merge_description"`
	MergeDateTaken       bool   `json:"mergeDateTaken" koanf:"merge_date_taken"`
	MergeURL             bool   `json:"mergeURL" koanf:"merge_URL"`
	MergeGPS             bool   `json:"mergeGPS" koanf:"merge_GPS"`
}

type GlobalSettingsT struct {
	MergeSettings MergeSettings `json:"mergeSettings" koanf:"merge_settings"`
	Window        struct {
		Saved       bool `json:"saved" koanf:"saved"`
		IsMaximised bool `json:"isMaximised" koanf:"is_maximised"`
		SizeW       int  `json:"sizeW" koanf:"size_w"`
		SizeH       int  `json:"sizeH" koanf:"size_h"`
		PosX        int  `json:"posX" koanf:"pos_x"`
		PosY        int  `json:"posY" koanf:"pos_y"`
	} `json:"window" koanf:"window"`
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
	GlobalSettings.MergeSettings.MergeTitle = true
	GlobalSettings.MergeSettings.MergeDescription = true
	GlobalSettings.MergeSettings.MergeDateTaken = true
	GlobalSettings.MergeSettings.MergeURL = true
	GlobalSettings.MergeSettings.MergeGPS = true
	GlobalSettings.MergeSettings.TimezoneOffset = time.Now().Format("-0700")
	fmt.Println(GlobalSettings)
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

package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/tidwall/gjson"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/wailsapp/wails/v3/pkg/application"
)

var GlobalSettingsConfig GlobalSettins
var UserSettingsDir string = GetUserDir() + "/.config/go-out"
var GlobalSettingsPath string = GetUserDir() + "/.config/go-out/config.yaml"

type GlobalSettins struct {
	Window struct {
		Saved       bool `json:"saved" koanf:"saved"`
		IsMaximised bool `json:"isMaximised" koanf:"is_maximised"`
		SizeW       int  `json:"sizeW" koanf:"size_w"`
		SizeH       int  `json:"sizeH" koanf:"size_h"`
		PosX        int  `json:"posX" koanf:"pos_x"`
		PosY        int  `json:"posY" koanf:"pos_y"`
	} `json:"window" koanf:"window"`
}

//go:embed build/info.json
var WailsInfoJSON string

func RestoreSettings(app *application.App, window *application.WebviewWindow) {
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
}

func GetAppVersion() string {
	return gjson.Get(WailsInfoJSON, "info.0000.ProductVersion").String()
}

func GetUserDir() string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return dirname
}

func ParseGlobalConfig() (GlobalSettins, error) {
	var c GlobalSettins
	var k = koanf.New(".")
	if err := k.Load(file.Provider(GlobalSettingsPath), yaml.Parser()); err != nil {
		log.Printf("error loading global config: %v", err)
		return GlobalSettins{}, err
	}
	err := k.Unmarshal("", &c)
	if err != nil {
		log.Printf("error Unmarshaling global config: %v", err)
		return GlobalSettins{}, err
	}

	return c, nil
}

func MakeNewDefaultConfig() error {
	GlobalSettingsConfig = GlobalSettins{}
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

	err := k.Load(structs.Provider(GlobalSettingsConfig, "koanf"), nil)
	if err != nil {
		fmt.Println(err)
		return err
	}
	b, err := k.Marshal(yaml.Parser())
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = os.WriteFile(GlobalSettingsPath, b, os.ModePerm)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

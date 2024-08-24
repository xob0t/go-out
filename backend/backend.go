package backend

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/tidwall/gjson"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func GetAllJsonFiles(paths []string) ([]string, error) {
	var jsons []string

	for _, root := range paths {
		info, err := os.Stat(root)
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			// If the path is a directory, walk through it
			err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				// If it's a file and has a .json extension, add it to the list
				if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
					jsons = append(jsons, path)
				}

				return nil
			})
			if err != nil {
				return nil, err
			}
		} else if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
			// If the path is a file and has a .json extension, add it to the list
			jsons = append(jsons, root)
		}
	}

	return jsons, nil
}

// UpdateMetadata updates the geo data for the image files based on the JSON metadata.
func UpdateMetadata(app *application.App, jsonPaths []string, editedSuffix string, processEdited bool, ignoreMinorErrors bool) {
	// Create a map to group JSON files and their corresponding image files.
	fileMap := make(map[string][]string)

	for _, jsonPath := range jsonPaths {
		// Extract the base name without the .json extension.
		mediaPath := strings.TrimSuffix(jsonPath, ".json")
		mediaExt := filepath.Ext(mediaPath)
		mediaPathNoExt := strings.TrimSuffix(mediaPath, mediaExt)
		editedName := mediaPathNoExt + editedSuffix + mediaExt

		// Check if the original media file exists.
		if _, err := os.Stat(mediaPath); err == nil {
			fileMap[jsonPath] = append(fileMap[jsonPath], mediaPath)
		}

		// If processEdited is true, check if the edited version exists.
		if processEdited {
			if _, err := os.Stat(editedName); err == nil {
				fileMap[jsonPath] = append(fileMap[jsonPath], editedName)
			}
		}
	}

	// Initialize slice of configuration functions
	var configFuncs []func(*exiftool.Exiftool) error

	// Add IgnoreMinorErrors function if required
	if ignoreMinorErrors {
		configFuncs = append(configFuncs, exiftool.IgnoreMinorErrors())
	}

	// Initialize exiftool with the specified configuration functions
	et, err := exiftool.NewExiftool(configFuncs...)
	if err != nil {
		errMsg := "Error initializing Exiftool: " + err.Error()
		app.Events.Emit(&application.WailsEvent{
			Name: "log",
			Data: map[string]string{
				"level":   "ERROR",
				"message": errMsg,
			},
		})
		app.Logger.Error(errMsg)
	}
	defer et.Close()

	// Process each JSON and image pair
	for jsonPath, imagePaths := range fileMap {
		// Read the JSON file
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			errMsg := "Error reading JSON file " + jsonPath + " " + err.Error()
			app.Events.Emit(&application.WailsEvent{
				Name: "log",
				Data: map[string]string{
					"level":   "ERROR",
					"message": errMsg,
				},
			})
			app.Logger.Error(errMsg)
			continue
		}

		// Use gjson to extract values
		title := gjson.GetBytes(data, "title").String()
		description := gjson.GetBytes(data, "description").String()
		photoTakenTime := gjson.GetBytes(data, "photoTakenTime.timestamp").Int()
		url := gjson.GetBytes(data, "url").String()
		latitude := gjson.GetBytes(data, "geoData.latitude").Float()
		longitude := gjson.GetBytes(data, "geoData.longitude").Float()
		altitude := gjson.GetBytes(data, "geoData.altitude").Float()

		t := time.Unix(photoTakenTime, 0).UTC()
		// Format the time as "YYYY:MM:DD HH:MM:SS"
		formattedTime := t.Format("2006:01:02 15:04:05")

		// Prepare the fields for GPS metadata for each corresponding image
		for _, imagePath := range imagePaths {
			fileMetadataSlice := []exiftool.FileMetadata{
				{
					File: imagePath,
					Fields: map[string]interface{}{
						"Title":            title,
						"URL":              url,
						"DateTimeOriginal": formattedTime,
						"ImageDescription": description,
					},
				},
			}

			// Add GPS data if present
			if latitude != 0 {
				fileMetadataSlice[0].Fields["GPSLatitude"] = latitude
			}
			if longitude != 0 {
				fileMetadataSlice[0].Fields["GPSLongitude"] = longitude
			}
			if altitude != 0 {
				fileMetadataSlice[0].Fields["GPSAltitude"] = altitude
			}

			// Write the metadata to the image
			et.WriteMetadata(fileMetadataSlice)

			// Check if there were any errors
			if fileMetadataSlice[0].Err != nil {
				errMsg := "Error writing data to image " + imagePath + ": " + fileMetadataSlice[0].Err.Error()
				app.Events.Emit(&application.WailsEvent{
					Name: "log",
					Data: map[string]string{
						"level":   "ERROR",
						"message": errMsg,
					},
				})
				app.Logger.Error(errMsg)
				continue
			}

			logMsg := "Successfully updated EXIF data for image: " + imagePath
			app.Events.Emit(&application.WailsEvent{
				Name: "log",
				Data: map[string]string{
					"level":   "INFO",
					"message": logMsg,
				},
			})
			app.Logger.Info(logMsg)
		}
	}
}

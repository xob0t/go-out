package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/bradfitz/latlong"
	"github.com/tidwall/gjson"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func LogWrapper(app *application.App, level, message string) {
	// logs to both the UI and the console
	app.Events.Emit(&application.WailsEvent{
		Name: "log",
		Data: map[string]string{
			"level":   level,
			"message": message,
		},
	})
	switch level {
	case "INFO":
		app.Logger.Info(message)
	case "DEBUG":
		app.Logger.Debug(message)
	case "WARNING":
		app.Logger.Warn(message)
	case "ERROR":
		app.Logger.Error(message)
	default:
		fmt.Println("Invalid level")
	}
}

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

func ApplyTimezoneOffset(t time.Time, offset string) time.Time {
	// Parse the offset
	offset = strings.TrimPrefix(offset, "+")
	offset = strings.TrimPrefix(offset, "-")
	hours, _ := strconv.Atoi(offset[:2])
	minutes, _ := strconv.Atoi(offset[2:])
	if strings.HasPrefix(offset, "-") {
		hours = -hours
		minutes = -minutes
	}

	// Create a fixed zone with the parsed offset
	location := time.FixedZone("Custom", hours*3600+minutes*60)

	// Apply the timezone offset to the time
	return t.In(location)
}

func getTimeInTimezone(t time.Time, lat, lon float64) (time.Time, error) {
	// Get timezone name from coordinates
	tz := latlong.LookupZoneName(lat, lon)
	if tz == "" {
		return time.Time{}, fmt.Errorf("unable to determine timezone for coordinates: %f, %f", lat, lon)
	}

	// Load the timezone location
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Time{}, fmt.Errorf("error loading timezone location: %v", err)
	}

	// Convert UTC time to the specified location's time
	localTime := t.In(loc)

	return localTime, nil
}

// UpdateMetadata updates file exif data based on the JSON metadata.
func UpdateMetadata(app *application.App, jsonPaths []string, mergeSettings MergeSettings) {
	// Create a map to group JSON files and their corresponding files.
	fileMap := make(map[string][]string)

	for _, jsonPath := range jsonPaths {
		// Extract the base name without the .json extension.
		mediaPath := strings.TrimSuffix(jsonPath, ".json")
		mediaExt := filepath.Ext(mediaPath)
		mediaPathNoExt := strings.TrimSuffix(mediaPath, mediaExt)
		editedName := mediaPathNoExt + "-" + mergeSettings.EditedSuffix + mediaExt

		// Check if the original media file exists.
		if _, err := os.Stat(mediaPath); err == nil {
			fileMap[jsonPath] = append(fileMap[jsonPath], mediaPath)
		}

		// If processEdited is true, check if the edited version exists.
		if _, err := os.Stat(editedName); err == nil {
			fileMap[jsonPath] = append(fileMap[jsonPath], editedName)
		}
	}

	// Initialize slice of configuration functions
	var configFuncs []func(*exiftool.Exiftool) error

	// Add IgnoreMinorErrors function if required
	if mergeSettings.IgnoreMinorErrors {
		configFuncs = append(configFuncs, exiftool.IgnoreMinorErrors())
	}

	// Initialize exiftool with the specified configuration functions
	et, err := exiftool.NewExiftool(configFuncs...)
	if err != nil {
		logMsg := "Error initializing Exiftool: " + err.Error()
		LogWrapper(app, "ERROR", logMsg)
	}
	defer et.Close()

	// Process each JSON and file pair
	for jsonPath, filePaths := range fileMap {
		// Read the JSON file
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			logMsg := "Error reading JSON file " + jsonPath + " " + err.Error()
			LogWrapper(app, "ERROR", logMsg)
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

		// Prepare the fields for exif metadata for each corresponding file
		for _, filePath := range filePaths {
			fileMetadataSlice := []exiftool.FileMetadata{
				{
					File:   filePath,
					Fields: map[string]interface{}{},
				},
			}
			if mergeSettings.MergeTitle {
				fileMetadataSlice[0].Fields["Title"] = title
			}
			if mergeSettings.MergeURL {
				fileMetadataSlice[0].Fields["URL"] = url
			}
			if mergeSettings.MergeDateTaken {
				t := time.Unix(photoTakenTime, 0).UTC()
				if latitude > 0 && longitude > 0 && mergeSettings.InferTimezoneFromGPS {
					t, err = getTimeInTimezone(t, latitude, longitude)
					if err != nil {
						LogWrapper(app, "ERROR", err.Error())
					}
				} else {
					t = ApplyTimezoneOffset(t, mergeSettings.TimezoneOffset)
				}
				// Format the time as "YYYY:MM:DD HH:MM:SS"
				formattedTime := t.Format("2006:01:02 15:04:05")
				fileMetadataSlice[0].Fields["DateTimeOriginal"] = formattedTime
			}
			if mergeSettings.MergeDescription {
				fileMetadataSlice[0].Fields["ImageDescription"] = description
			}

			if mergeSettings.MergeGPS {
				if latitude != 0 {
					fileMetadataSlice[0].Fields["GPSLatitude"] = latitude
				}
				if longitude != 0 {
					fileMetadataSlice[0].Fields["GPSLongitude"] = longitude
				}
				if altitude != 0 {
					fileMetadataSlice[0].Fields["GPSAltitude"] = altitude
				}
			}

			{
				logMsg := fmt.Sprint("Writing metadata: ", fileMetadataSlice[0].Fields)
				LogWrapper(app, "INfO", logMsg)
			}

			// Write the metadata to the file
			et.WriteMetadata(fileMetadataSlice)

			// Check if there were any errors
			if fileMetadataSlice[0].Err != nil {
				logMsg := "Error writing data to file " + filePath + ": " + fileMetadataSlice[0].Err.Error()
				LogWrapper(app, "ERROR", logMsg)
				continue
			}

			logMsg := "Successfully updated EXIF data for file: " + filePath
			LogWrapper(app, "INFO", logMsg)
		}
	}
}

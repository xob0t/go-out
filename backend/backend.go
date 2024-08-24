package backend

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/tidwall/gjson"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func GetAllFiles(root string) ([]string, error) {
	var files []string

	// Walk through the directory tree recursively
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// If it's a file (not a directory), add it to the list
		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// GeoData struct represents the geolocation data in the JSON file.
type GeoData struct {
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Altitude      float64 `json:"altitude"`
	LatitudeSpan  float64 `json:"latitudeSpan"`
	LongitudeSpan float64 `json:"longitudeSpan"`
}

// PhotoMetadata struct represents the overall structure of the JSON metadata.
type PhotoMetadata struct {
	GeoData GeoData `json:"geoData"`
}

// updateGeoData updates the geo data for the image files based on the JSON metadata.
func UpdateGeoData(app *application.App, filePaths []string, stringEdited string, processEdited bool) {
	// Create a map to group JSON files and their corresponding image files.

	app.Events.Emit(&application.WailsEvent{
		Name: "log",
		Data: map[string]string{
			"level":   "info",
			"message": "HELLO FROM BACKEND!",
		},
	})

	fileMap := make(map[string][]string)

	for _, filePath := range filePaths {
		// Check if the file is a JSON file.
		if strings.HasSuffix(filePath, ".json") {
			// Extract the base name without the .json extension.
			mediaPath := strings.TrimSuffix(filePath, ".json")
			mediaExt := filepath.Ext(mediaPath)
			mediaPathNoExt := strings.TrimSuffix(mediaPath, mediaExt)
			editedName := mediaPathNoExt + stringEdited + mediaExt

			// Search for the corresponding media
			for _, imgPath := range filePaths {
				if imgPath == mediaPath || (processEdited && imgPath == editedName) {
					fileMap[filePath] = append(fileMap[filePath], imgPath)
				}
			}
		}
	}

	// Initialize exiftool
	et, err := exiftool.NewExiftool(exiftool.IgnoreMinorErrors())
	if err != nil {
		log.Fatalf("Error initializing Exiftool: %v", err)
	}
	defer et.Close()

	// Process each JSON and image pair
	for jsonPath, imagePaths := range fileMap {
		// Read the JSON file
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			log.Printf("Error reading JSON file %s: %v", jsonPath, err)
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

		fmt.Println(fmt.Sprint(title+"\n", description+"\n", url+"\n", formattedTime+"\n"))
		// Debugging output
		fmt.Printf("Extracted Data from %s: Latitude: %f, Longitude: %f, Altitude: %f\n", jsonPath, latitude, longitude, altitude)
		fmt.Println()

		// Prepare the fields for GPS metadata for each corresponding image
		for _, imagePath := range imagePaths {

			// original := et.ExtractMetadata(imagePath)

			fileMetadataSlice := []exiftool.FileMetadata{
				{
					File: imagePath,
					Fields: map[string]interface{}{
						"Title":            "fooxqa",
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
				log.Printf("Error writing GPS data to image %s: %v", imagePath, fileMetadataSlice[0].Err)
				continue
			}

			fmt.Println("\nSuccessfully updated GPS data for image: ", imagePath)

			modified := et.ExtractMetadata(imagePath)
			if modified[0].Fields["Title"] != "fooxqa" {
				fmt.Println("error setting title!")
			}
		}
	}
}

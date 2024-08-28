import { SettingsService } from "./bindings/go-out";

export function initializeSettings() {
  document.querySelectorAll("input").forEach((input) => {
    input.addEventListener("change", updateInputState);
  });

  let settingsState = {
    exifTags: {}
  };

  SettingsService.Get()
    .then((result) => {
      settingsState = result;
      console.log("settings received", settingsState);
      initializeInputs(settingsState);
    })
    .catch((err) => {
      console.error("settings get err", err);
    });

  function initializeInputs(settingsState) {
    // Initialize settings fields
    const generalSettings = ["editedSuffix", "ignoreMinorErrors", "timezoneOffset", "inferTimezoneFromGPS", "overwriteExistingTags"];
    generalSettings.forEach((inputId) => {
      const inputElement = document.getElementById(inputId);
      if (inputElement) {
        if (inputElement.type === "checkbox") {
          inputElement.checked = settingsState[inputId];
        } else {
          inputElement.value = settingsState[inputId];
        }
      }
    });

    // Initialize ExifTags fields
    const exifTags = ["title", "description", "dateTaken", "URL", "GPS"];
    exifTags.forEach((inputId) => {
      const inputElement = document.getElementById(inputId);
      if (inputElement) {
        inputElement.checked = settingsState.exifTags[inputId];
      }
    });
  }

  function updateInputState(event) {
    const inputId = event.target.id;

    if (inputId in settingsState.exifTags) {
      // Update ExifTags fields
      settingsState.exifTags[inputId] = event.target.checked;
    } else {
      // Update general settings fields
      if (event.target.type === "checkbox") {
        settingsState[inputId] = event.target.checked;
      } else {
        settingsState[inputId] = event.target.value;
      }
    }

    SettingsService.Update(settingsState)
      .then(() => {
        console.log("settings updated");
      })
      .catch((err) => {
        console.error("settings update err", err);
      });
  }
}

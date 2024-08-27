import { SettingsService } from "./bindings/go-out";

export function initializeSettings() {
  document.querySelectorAll("input").forEach((input) => {
    input.addEventListener("change", updateInputState);
  });

  let settingsState = {};

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
    for (let inputId in settingsState) {
      const inputElement = document.getElementById(inputId);
      if (inputElement) {
        if (inputElement.type === "checkbox") {
          inputElement.checked = settingsState[inputId];
        } else {
          inputElement.value = settingsState[inputId];
        }
      }
    }
  }

  function updateInputState(event) {
    const inputId = event.target.id;
    if (event.target.type === "checkbox") {
      settingsState[inputId] = event.target.checked;
    } else {
      settingsState[inputId] = event.target.value;
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

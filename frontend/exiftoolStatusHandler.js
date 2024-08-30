import { Events, Browser } from "@wailsio/runtime";

function displayWarinig() {
  const overlay = document.getElementById("warning-overlay");
  const warning = document.getElementById("warning-content");
  const app = document.getElementById("app-main");
  app.style.display = "none";
  overlay.style.display = "flex";

  warning.addEventListener("click", function () {
    Browser.OpenURL("https://exiftool.org/");
  });
}

export function exiftoolStatusHandler() {
  Events.On("exiftoolStatus", (event) => {
    const exiftoolStatus = event.data;
    console.log("exiftoolStatus", exiftoolStatus);
    if (!exiftoolStatus) {
      displayWarinig();
    }
  });
}

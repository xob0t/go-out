import { initializeSettings } from "./settings";
import { initializeLogArea } from "./logArea";
import { exiftoolStatusHandler } from "./exiftoolStatusHandler";

exiftoolStatusHandler();
initializeSettings();
initializeLogArea();

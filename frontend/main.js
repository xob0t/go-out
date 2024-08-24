import { Events } from "@wailsio/runtime";

// Array to store log history
const logHistory = [];

// Function to update the log area with the latest log entry
function updateLogArea() {
  const logArea = document.getElementById("logArea");
  logArea.innerHTML = ""; // Clear existing content
  if (logHistory.length > 0) {
    const latestLog = logHistory[logHistory.length - 1];
    const logEntry = document.createElement("div");
    logEntry.className = `log-entry ${latestLog.level}`;
    logEntry.textContent = `${new Date(latestLog.timestamp).toLocaleTimeString()} [${latestLog.level.toUpperCase()}] ${latestLog.message}`;
    logArea.appendChild(logEntry);
  }
}

// Function to append log messages to the history array
function addLogToHistory(level, message) {
  const timestamp = new Date();
  logHistory.push({ level, message, timestamp });
  // Update the log area with the latest log entry
  updateLogArea();
  document.getElementById("clearLog").style.display = "block";
  document.getElementById("copyLog").style.display = "block";
}

// Function to toggle the full log display
function toggleLogHistory() {
  const logArea = document.getElementById("logArea");
  const toggleButton = document.getElementById("toggleLogVisibility");
  const isExpanded = logArea.classList.contains("expanded");

  if (!isExpanded) {
    // Expand and show full log history
    logArea.classList.add("expanded");
    logArea.innerHTML = "";
    logHistory.forEach((log) => {
      const logEntry = document.createElement("div");
      logEntry.className = `log-entry ${log.level}`;
      logEntry.textContent = `${new Date(log.timestamp).toLocaleTimeString()} [${log.level.toUpperCase()}] ${log.message}`;
      logArea.appendChild(logEntry);
    });
    // Show the "Hide Log" button
    toggleButton.style.display = "block";
  }
}

// Function to collapse and show only the latest log
function collapseLog() {
  const logArea = document.getElementById("logArea");
  const toggleButton = document.getElementById("toggleLogVisibility");
  logArea.classList.remove("expanded");
  updateLogArea();
  // Hide the "Hide Log" button
  toggleButton.style.display = "none";
}

// Set up the event listeners
document.getElementById("logArea").addEventListener("click", function () {
  if (!this.classList.contains("expanded")) {
    toggleLogHistory();
  }
});

document.getElementById("clearLog").addEventListener("click", function () {
  const logArea = document.getElementById("logArea");
  logArea.innerHTML = "";
  this.style.display = "none";
  const toggleButton = document.getElementById("toggleLogVisibility");
  toggleButton.style.display = "none";
  const copyLogButton = document.getElementById("copyLog");
  copyLogButton.style.display = "none";
});

document.getElementById("copyLog").addEventListener("click", function () {
  // Create a string from all log messages
  const logText = logHistory.map((log) => `${new Date(log.timestamp).toLocaleTimeString()} [${log.level.toUpperCase()}] ${log.message}`).join("\n");

  // Copy the logText to the clipboard
  navigator.clipboard
    .writeText(logText)
    .then(() => {
      console.log("Log copied to clipboard!");
    })
    .catch((err) => {
      console.error("Failed to copy log: ", err);
    });
    
});

document.getElementById("toggleLogVisibility").addEventListener("click", collapseLog);

// Listen for 'log' events emitted from the Go backend
Events.On("log", (event) => {
  console.log("log event received", event.data);
  console.log(event.data.level);
  console.log(event.data.message);
  addLogToHistory(event.data.level, event.data.message);
});
Events.On("time", (time) => {
  timeElement.innerText = time.data;
});

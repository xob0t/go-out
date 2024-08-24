import {Events} from "@wailsio/runtime";

// Function to append log messages to the log area
function appendLog(level, message) {
    const logArea = document.getElementById('logArea');
    const logEntry = document.createElement('div');
    logEntry.className = `log-entry ${level}`;
    logEntry.textContent = `${new Date().toLocaleTimeString()} [${level.toUpperCase()}] ${message}`;
    logArea.appendChild(logEntry);
}

// Listen for 'log' events emitted from the Go backend
Events.On('log', (event) => {
    console.log("log event recived", event.data)
    console.log(event.data.level)
    console.log(event.data.message)
    appendLog(event.data.level, event.data.message);

});

Events.On('time', (time) => {
    timeElement.innerText = time.data;
});

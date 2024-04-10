// Define a data structure to hold websocket connections and messages
class SocketConnections {
    constructor() {
        this.connections = {}; // Object to store websocket connections
        this.messages = {};    // Object to store messages associated with each connection
        this.activeConnection = "";
    }
  
    // Method to connect to a websocket
    connect(guid, myid) {
        if (!(guid in this.connections)) {
            let ws = new WebSocket(`ws://${chatHost}:${chatPort}/ws?guid=${guid}&userid=${myid}`);
            this.connections[guid] = ws;
            let messages = sse.getMessagesFromGuid(guid);
            this.messages[guid] = messages;
        }
        this.activeConnection = guid;
        let ws = this.connections[guid];
        this.renderMessagesToChatInterface(guid);
        
        ws.onmessage = (event) => {
        this.messages[guid].push(event.data); // Store the incoming message
        //append to the chat interface if this is the active connection
        if (this.activeConnection === guid) {
            this.renderSingleMessageToChatInterface(event.data);
        }
        };

        ws.onerror = (error) => {
        console.error("WebSocket error:", error);
        delete this.connections[guid]; // Remove the connection from the connections object
        delete this.messages[guid];    // Remove messages associated with this connection
        if (this.activeConnection === guid) {
            this.activeConnection = "";
        }
        };

        ws.onclose = () => {
        delete this.connections[guid]; // Remove the connection from the connections object
        delete this.messages[guid];    // Remove messages associated with this connection
        if (this.activeConnection === guid) {
            this.activeConnection = "";
        }
        }; 
    }
  
    // Method to send a message through a websocket connection
    sendMessage(message) {
        const ws = this.connections[this.activeConnection];
        if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(message);
        };
    }
  
    // Method to close a websocket connection
    closeConnection(guid) {
        const ws = this.connections[guid];
        let displayLoc = document.getElementById("right-chat-body");
        displayLoc.innerHTML = "";
        this.activeConnection = "";
        if (ws && ws.readyState !== WebSocket.CLOSED) {
        ws.close();
        return true; // Connection closed successfully
        }
        return false; // Connection already closed or not found
    }

    renderSingleMessageToChatInterface(message) {
        let displayLoc = document.getElementById("right-chat-body");
        displayLoc.appendChild(createTileDiv(message, ""));
    }
    

    renderMessagesToChatInterface(guid) {
        let displayLoc = document.getElementById("right-chat-body");
        displayLoc.innerHTML = "";
        this.messages[guid].forEach(message => {
            displayLoc.appendChild(createTileDiv(message, ""));
        });
        showleaveButton(guid);
    }
    

    setActiveConnection(guid) {
        if (guid === "") {

        }
        this.activeConnection = guid;
    }

    getActiveConnection() {
        return this.activeConnection;
    }

    //messageArray should be in the format of [name: message, name: message, ...], getmessagedFromGuid in the sseEvents.js file
    addMessagesFromArray(guid, messageArray) {
        messageArray.forEach(message => {
            this.messages[guid].push(message);
        });
    }


}


// Function to create a tile div for a chat participant
function createTileDiv(nameMessage, avatarSrc) {

    //special case for user join / leave messages
    if (!nameMessage.includes(":")) {
        let tileDiv = document.createElement('div');
        tileDiv.classList.add('tile');
        let tileContentDiv = document.createElement('div');
        tileContentDiv.classList.add('tile-content');
        let tileSubtitle = document.createElement('p');
        tileSubtitle.classList.add('tile-subtitle');
        tileSubtitle.textContent = nameMessage;
        tileContentDiv.appendChild(tileSubtitle);
        tileDiv.appendChild(tileContentDiv);
        return tileDiv
    }

    let name = nameMessage.split(":")[0];
    let subtitle = nameMessage.split(":")[1].trim();

    let tileDiv = document.createElement('div');
    tileDiv.classList.add('tile');

    let tileIconDiv = document.createElement('div');
    tileIconDiv.classList.add('tile-icon');

    let avatarFigure = document.createElement('figure');
    avatarFigure.classList.add('avatar');
    if (avatarSrc === "") {
        avatarFigure.setAttribute('data-initial', name.charAt(0).toUpperCase());
    }
    let avatarImg = document.createElement('img');
    avatarImg.src = avatarSrc;
    avatarFigure.appendChild(avatarImg);
    tileIconDiv.appendChild(avatarFigure);

    let tileContentDiv = document.createElement('div');
    tileContentDiv.classList.add('tile-content');

    let tileTitle = document.createElement('p');
    tileTitle.classList.add('tile-title', 'text-bold');
    tileTitle.textContent = name;
    tileContentDiv.appendChild(tileTitle);

    let tileSubtitle = document.createElement('p');
    tileSubtitle.classList.add('tile-subtitle');
    tileSubtitle.textContent = subtitle;
    tileContentDiv.appendChild(tileSubtitle);

    tileDiv.appendChild(tileIconDiv);
    tileDiv.appendChild(tileContentDiv);

    return tileDiv;
}

function showleaveButton(guid) {
    let leaveButton = document.getElementById("leave");
    leaveButton.style.display = "block";
    leaveButton.onclick = function() {
        sockets.closeConnection();
        leaveButton.style.display = "none";
    }
}
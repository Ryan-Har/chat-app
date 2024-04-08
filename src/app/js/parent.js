document.addEventListener("DOMContentLoaded", function() {
    const availableChatsButton = document.getElementById("available-chats-button");
    const myChatsButton = document.getElementById("my-chats-button");
    const logoutButton = document.getElementById("logout-button");

    availableChatsButton.addEventListener("click", function(event) {
        page = "availableChats";
        loadChildPageContent(event, "/chats");
    });
    
    myChatsButton.addEventListener("click", function(event) {
        page = "myChats";
        loadChildPageContent(event, "/chats");
    });

    // Get the navbar element
    var navbar = document.getElementById("navbar");
    // Get the computed height of the navbar
    var navbarHeight = navbar.offsetHeight;
    // Set the height of the container dynamically
    var container = document.querySelector(".container");
    container.style.height = "calc(100% - " + navbarHeight + "px)";
});

let availableChats = [];
let myChats = [];
let allChats = [];
let myid= 1;
let page = "";

const eventSource = new EventSource("/chatstream");
    eventSource.onmessage = (event) => {
    const message = JSON.parse(event.data);
    //ogranise into chats owned by me, chats owned by others and chats that are owned by noone
    allChats = "";
    allChats = message;
    myChats = message.filter(chat => hasMeInChat(chat, myid));
    availableChats = message.filter(chat => freeForPickup(chat, myid));
    handleLeftChatBody();
};

function hasMeInChat(chatArray, myid) {
    for (const participant of chatArray.participants) {
      if (participant.userid === myid && participant.active === true) {
        return true;
      }
    }
  return false;
}

function freeForPickup(chatArray, myid) {
    for (const participant of chatArray.participants) {
      if (participant.internal === true && participant.active === true) {
        return false;
      }
    }
  return true;
}

function loadChildPageContent(event, pageUrl) {
    event.preventDefault();
    console.log("loading child page content");
    // Make an AJAX request to fetch content from child page
    let xhr = new XMLHttpRequest();
    xhr.open("GET", pageUrl, true);
    xhr.onreadystatechange = function() {
        if (xhr.readyState === 4 && xhr.status === 200) {

            // Insert fetched content into parent page
            document.getElementById("content").innerHTML = xhr.responseText;
            if (pageUrl === "/chats") {
                handleLeftChatBody();
            }
        }
    };
    xhr.send();

}

function handleLeftChatBody() {
    if (page === "availableChats") {
        displayAllChats();
        //displayAvailableChats();
    } else if (page === "myChats" && myChats.length > 0) {
        displayMyChats();
    }
}

function displayAvailableChats() {
    console.log("displaying available chats");
    const messageMaxChars = 20;
    let displayLoc = document.getElementById("left-chat-body");
    displayLoc.innerHTML = "";
    availableChats.forEach(chat => {
        let tile = document.createElement("div");
        tile.classList.add("tile", "tile-centered", "chat-tile");

        let tileContent = document.createElement("div");
        tileContent.classList.add("tile-content");

        let tileTitle = document.createElement("div");
        tileTitle.classList.add("tile-title");
        for (const participant of chat.participants) {
            if (participant.internal === false && participant.active === true) {
                tileTitle.textContent = participant.name;
            }
        }
        let tileTitleSpan = document.createElement("span");
        tileTitleSpan.textContent = chat.chatStartTime;
        tileTitle.appendChild(tileTitleSpan);

        let tileSubtitle = document.createElement("small");
        tileSubtitle.classList.add("tile-subtitle");
        
        // Truncate the message string if it exceeds maxChars
        let truncatedMessage = chat.messages[0].message.slice(0, messageMaxChars);
        if (chat.messages[0].message.length > messageMaxChars) {
            truncatedMessage += "...";
        }
        tileSubtitle.textContent = truncatedMessage;

        tileContent.appendChild(tileTitle);
        tileContent.appendChild(tileSubtitle);
        tile.appendChild(tileContent);
        displayLoc.appendChild(tile);
        console.log("displaying available chats");
    });
}

function displayMyChats() {
    console.log("displaying my chats");
}

function displayAllChats() {
    console.log("displaying all chats");
    console.log("displaying available chats");
    const messageMaxChars = 20;
    let displayLoc = document.getElementById("left-chat-body");
    displayLoc.innerHTML = "";
    allChats.forEach(chat => {
        let tile = document.createElement("div");
        tile.classList.add("tile", "tile-centered", "chat-tile");

        let tileContent = document.createElement("div");
        tileContent.classList.add("tile-content");

        let tileTitle = document.createElement("div");
        tileTitle.classList.add("tile-title");
        for (const participant of chat.participants) {
            if (participant.internal === false && participant.active === true) {
                tileTitle.textContent = participant.name;
            }
        }
        let tileTitleSpan = document.createElement("span");
        tileTitleSpan.textContent = chat.chatStartTime;
        tileTitle.appendChild(tileTitleSpan);

        let tileSubtitle = document.createElement("small");
        tileSubtitle.classList.add("tile-subtitle");
        
        // Truncate the message string if it exceeds maxChars
        let truncatedMessage = chat.messages[0].message.slice(0, messageMaxChars);
        if (chat.messages[0].message.length > messageMaxChars) {
            truncatedMessage += "...";
        }
        tileSubtitle.textContent = truncatedMessage;

        let tileAction = document.createElement("div");
        tileAction.classList.add("tile-action");
        chat.participants.forEach(participant => {
            if (participant.internal === true && participant.active === true) {
                let chip = document.createElement("div");
                chip.classList.add("chip");
                let chipFigure = document.createElement("figure");
                chipFigure.classList.add("avatar", "avatar-sm"),
                chipFigure.setAttribute("data-initial", participant.name.charAt(0).toUpperCase());
                let chipIcon = document.createElement("img");
                chipIcon.src = "";
                chipFigure.appendChild(chipIcon);     
                chip.appendChild(chipFigure);
                let chipText = document.createTextNode(participant.name);
                chip.appendChild(chipText);
                tileAction.appendChild(chip);
            }
        });
        
        tileContent.appendChild(tileTitle);
        tileContent.appendChild(tileSubtitle);
        tileContent.appendChild(tileAction);
        tile.appendChild(tileContent);
        displayLoc.appendChild(tile);
        console.log("displaying available chats");
    });
}
//   let ws;


// const chatMessages = document.getElementById("chat-messages");

// function connectToChat(guid) {
//       console.log(guid)
//       lastConnectedGuid = guid
//       return new Promise((resolve, reject) => {
//           const internalUserID = 1; 
  
//           // WebSocket connection URL with the GUID and user's name
//           const wsUrl = `ws://${chatHost}:${chatPort}/ws?guid=${guid}&userid=${internalUserID.toString()}`;
//           console.log(wsUrl);
  
//           // Establish WebSocket connection
//           ws = new WebSocket(wsUrl);
  
//           // Set up event listeners for WebSocket
//           ws.onopen = () => {
//               console.log("WebSocket connection established");
//               resolve();
//           };
  
//           ws.onmessage = (event) => {
//               let newMessage = document.createElement('div');
//               newMessage.className = 'message';
//               newMessage.textContent = event.data;
  
//               chatMessages.appendChild(newMessage);
//           };
  
//           ws.onclose = () => {
//               console.log("WebSocket connection closed");
//               chatMessages.innerHTML = "";
//           };
  
//           ws.onerror = (error) => {
//               console.error("WebSocket error:", error);
//               reject(error);
//           };
//       });
//   }
  
//   let lastConnectedGuid;
  
//   async function sendMessage(guid) {
//       const messageInput = document.getElementById("message-input");
//       const message = messageInput.value.trim();
//       if (arguments.length > 0 && (!ws || ws.readyState != WebSocket.OPEN)) {
//           await connectToChat(guid);
//       }
  
//       if (arguments.length > 0 && guid != lastConnectedGuid) {
//           ws.close();
//           await connectToChat(guid);
//       }
  
//       if (message && ws && ws.readyState === WebSocket.OPEN) {
//           // Send message to the server
//           ws.send(message);
//       }
        
//       messageInput.value = "";
//   }
  
//   const compareStartTimes = (a, b) => {
//       const dateA = new Date(a.starttime);
//       const dateB = new Date(b.starttime);
//       return dateA - dateB;
//     };


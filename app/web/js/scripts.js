/*!
    * Start Bootstrap - SB Admin v7.0.7 (https://startbootstrap.com/template/sb-admin)
    * Copyright 2013-2023 Start Bootstrap
    * Licensed under MIT (https://github.com/StartBootstrap/startbootstrap-sb-admin/blob/master/LICENSE)
    */
    // 
// Scripts
// 

window.addEventListener('DOMContentLoaded', event => {

    // Toggle the side navigation
    const sidebarToggle = document.body.querySelector('#sidebarToggle');
    if (sidebarToggle) {
        // Uncomment Below to persist sidebar toggle between refreshes
        // if (localStorage.getItem('sb|sidebar-toggle') === 'true') {
        //     document.body.classList.toggle('sb-sidenav-toggled');
        // }
        sidebarToggle.addEventListener('click', event => {
            event.preventDefault();
            document.body.classList.toggle('sb-sidenav-toggled');
            localStorage.setItem('sb|sidebar-toggle', document.body.classList.contains('sb-sidenav-toggled'));
        });
    }

});

let messageInput = document.getElementById('message-input');
messageInput.addEventListener("keypress", function(event) {
    if (event.key === "Enter") {
      event.preventDefault();
      document.getElementById("send-button").click();
    }
  }); 

const eventSource = new EventSource("/chats");
eventSource.onmessage = (event) => {
  const message = JSON.parse(event.data);
  message.sort(compareStartTimes);
  const element = document.getElementById("ongoing-chats");
  element.innerHTML = '';
  message.forEach(chat => {
    element.innerHTML += 
    `<a class="nav-link" onclick="sendMessage('${chat.uuid}')" href="javascript:void(0);"></div>
    <div class="sb-nav-link-icon">
    
    <div class="row border">
    <div class="col-12">${chat.uuid}</div>
    <div class="col-12">${chat.starttime}</div>
    </div>
    </a>`;
  });
};

let ws;
const chatMessages = document.getElementById("chat-messages");
  
function connectToChat(guid) {
    console.log(guid)
    lastConnectedGuid = guid
    return new Promise((resolve, reject) => {
        const internalUserID = 1; 

        // WebSocket connection URL with the GUID and user's name
        const wsUrl = `ws://${chatHost}:${chatPort}/ws?guid=${guid}&userid=${internalUserID.toString()}`;
        console.log(wsUrl);

        // Establish WebSocket connection
        ws = new WebSocket(wsUrl);

        // Set up event listeners for WebSocket
        ws.onopen = () => {
            console.log("WebSocket connection established");
            resolve();
        };

        ws.onmessage = (event) => {
            let newMessage = document.createElement('div');
            newMessage.className = 'message';
            newMessage.textContent = event.data;

            chatMessages.appendChild(newMessage);
        };

        ws.onclose = () => {
            console.log("WebSocket connection closed");
            chatMessages.innerHTML = "";
        };

        ws.onerror = (error) => {
            console.error("WebSocket error:", error);
            reject(error);
        };
    });
}

let lastConnectedGuid;

async function sendMessage(guid) {
    const messageInput = document.getElementById("message-input");
    const message = messageInput.value.trim();
    if (arguments.length > 0 && (!ws || ws.readyState != WebSocket.OPEN)) {
        await connectToChat(guid);
    }

    if (arguments.length > 0 && guid != lastConnectedGuid) {
        ws.close();
        await connectToChat(guid);
    }

    if (message && ws && ws.readyState === WebSocket.OPEN) {
        // Send message to the server
        ws.send(message);
    }
      
    messageInput.value = "";
}

const compareStartTimes = (a, b) => {
    const dateA = new Date(a.starttime);
    const dateB = new Date(b.starttime);
    return dateA - dateB;
  };
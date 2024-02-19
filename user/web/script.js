let messageInput = document.getElementById('message-input');
messageInput.addEventListener("keypress", function(event) {
    if (event.key === "Enter") {
      event.preventDefault();
      document.getElementById("send-button").click();
    }
  }); 



let ws;
const chatMessages = document.getElementById("chat-messages");
const nameInput = document.getElementById("name-input");
  
function connectToChat() {
    return new Promise((resolve, reject) => {
        const name = nameInput.value.trim();
        if (!name) {
            alert("Please enter your name");
            return;
        }
        nameInput.style.visibility = "hidden";
        
        const guid = crypto.randomUUID();

        // WebSocket connection URL with the GUID and user's name
        const wsUrl = `ws://localhost:8002/ws?guid=${guid}&name=${name}`;

        // Establish WebSocket connection
        ws = new WebSocket(wsUrl);

        // Set up event listeners for WebSocket
        ws.onopen = () => {
            console.log("WebSocket connection established");
            resolve();run 
        };

        ws.onmessage = (event) => {
            let newMessage = document.createElement('div');
            newMessage.className = 'message';
            newMessage.textContent = event.data;

            chatMessages.appendChild(newMessage);
        };

        ws.onclose = () => {
            console.log("WebSocket connection closed");
        };

        ws.onerror = (error) => {
            console.error("WebSocket error:", error);
            reject(error);
        };
    });
}

async function sendMessage() {
    const messageInput = document.getElementById("message-input");
    const message = messageInput.value.trim();
    if (message && (!ws || ws.readyState != WebSocket.OPEN)) {
        await connectToChat();
    }

    if (message && ws && ws.readyState === WebSocket.OPEN) {
        // Send message to the server
        ws.send(message);
    }
      
    messageInput.value = "";
}




































  


// async function sendMessage() {
//     const messageInput = document.getElementById('message-input');
//     const message = messageInput.value;
//     if (message.trim() !== '') {
//         let chatMessages = document.getElementById('chat-messages');
//         const messageObject = { 'message': message };
//         const messageJsonString = JSON.stringify(messageObject);
                
//         if (chatMessages.childElementCount < 1) {
//             response = await fetch('http://localhost:8001/api/newmessage', {
//             method: 'POST',
//             body: messageJsonString,
//             headers: {
//                 'Content-type': 'application/json; charset=UTF-8'
//             }
//         }).then(response => response.json())
//         .then(data => {
//             console.log('POST request succeeded with JSON response', data);
//             newMessage = document.createElement('div');
//             newMessage.className = 'message';
//             newMessage.textContent = data.message;
//             chatMessages.appendChild(newMessage);
//         })
//         .catch(error => {
//             console.error('Error making POST request', error);
//         });
//         }

//         // var newMessage = document.createElement('div');
//         // newMessage.className = 'message';
//         // newMessage.textContent = message;

//         // chatMessages.appendChild(newMessage);

//         // You can implement server-side logic here to handle the message.
//         // For simplicity, we are just displaying the message on the client side.

//         //messageInput.value = '';
//     }
// }

// function connectToChat() {
//     const name = nameInput.value.trim();
//     if (!name) {
//         alert("Please enter your name");
//         return;
//     }

//     const guid = "yourUniqueChatRoomGuid";  // Replace with the actual GUID

//     // WebSocket connection URL with the GUID and user's name
//     const wsUrl = `ws://localhost:8000/ws?guid=${guid}&name=${name}`;

//     // Establish WebSocket connection
//     ws = new WebSocket(wsUrl);

//     // Set up event listeners for WebSocket
//     ws.onopen = () => {
//         console.log("WebSocket connection established");
//     };

//     ws.onmessage = (event) => {
//         const li = document.createElement("li");
//         li.textContent = event.data;
//         chatMessages.appendChild(li);
//     };

//     ws.onclose = () => {
//         console.log("WebSocket connection closed");
//     };

//     ws.onerror = (error) => {
//         console.error("WebSocket error:", error);
//     };
// }



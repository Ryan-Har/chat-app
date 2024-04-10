var chatSendButton = document.getElementById("chat-send-button");
var messageInput = document.getElementById("message-input");

chatSendButton.addEventListener("click", function() {
  console.log("Send button clicked");
  sendMessage();
  
});

messageInput.addEventListener("keypress", function(event) {
  if (event.key === "Enter") {
    console.log("Enter key pressed")
    event.preventDefault();
    document.getElementById("chat-send-button").click();
  }
}); 

function sendMessage() {
  let message = messageInput.value;
  let parentwindow = window.parent
  let dataToSend = {operation: "sendMessage", message: message};
  parentwindow.postMessage(dataToSend, "*");
  messageInput.value = "";
}

console.log("chat.js loaded")
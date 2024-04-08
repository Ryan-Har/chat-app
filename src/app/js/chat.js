const messageInput = document.getElementById('message-input');
messageInput.addEventListener("keypress", function(event) {
      if (event.key === "Enter") {
        event.preventDefault();
        document.getElementById("chat-send-button").click();
      }
}); 


console.log("chat.js loaded")
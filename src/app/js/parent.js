sse = new SSEvents();
sockets = new SocketConnections();

let myid= 1;
let currentPage = "";
let intervalId = null;

document.addEventListener("DOMContentLoaded", function() {
    const availableChatsButton = document.getElementById("available-chats-button");
    const myChatsButton = document.getElementById("my-chats-button");
    const allChatsButton = document.getElementById("all-chats-button");
    const logoutButton = document.getElementById("logout-button");

    availableChatsButton.addEventListener("click", function(event) {
        currentPage = "availableChats";
        loadChildPageContent(event, "/chats");
    });
    
    myChatsButton.addEventListener("click", function(event) {
        currentPage = "myChats";
        loadChildPageContent(event, "/chats");
    });

    allChatsButton.addEventListener("click", function(event) {
        currentPage = "allChats";
        loadChildPageContent(event, "/chats");
    });

    window.addEventListener("message", function(event) {
        let receivedData = event.data;
        console.log("received data from child page: ", receivedData)
        if (receivedData.operation === "sendMessage") {
            sockets.sendMessage(receivedData.message);
        }
    });

    // Get the navbar element
    var navbar = document.getElementById("navbar");
    // Get the computed height of the navbar
    var navbarHeight = navbar.offsetHeight;
    // Set the height of the container dynamically
    var container = document.querySelector(".container");
    container.style.height = "calc(100% - " + navbarHeight + "px)";
});


function loadChildPageContent(event, pageUrl) {
    event.preventDefault();
    console.log("loading child page content");
    // Make an AJAX request to fetch content from child page
    let xhr = new XMLHttpRequest();
    xhr.open("GET", pageUrl, true);
    xhr.onreadystatechange = function() {
        if (xhr.readyState === 4 && xhr.status === 200) {

            let htmlContent = xhr.responseText;
            let scriptTags = htmlContent.match(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi);
            if (scriptTags) {
                scriptTags.forEach(scriptTag => {
                    let scriptSrc = scriptTag.match(/src=["']([^"']+)["']/i)[1];
                    let script = document.createElement("script");
                    script.src = scriptSrc;
                    document.body.appendChild(script);
                });
            }
            // Insert fetched content into parent page
            document.getElementById("content").innerHTML = xhr.responseText;
            if (pageUrl === "/chats") {
                startIntervalForChildPage();
            }
        }
    };
    xhr.send();
}

function startIntervalForChildPage() {
    console.log("starting interval for child page");
    clearInterval(intervalId);
    console.log(currentPage)
    if (currentPage === "allChats") {
        intervalId = setInterval(function() {
            displayAllChats();
        }, 1000);
    } else if (currentPage === "myChats") {
        intervalId = setInterval(function() {
            displayMyChats();
        }, 1000);
    } else if (currentPage === "availableChats") {
        intervalId = setInterval(function() {
            displayAvailableChats();
        }, 1000);
    }
    reloadChatMessages();
}

function displayAvailableChats() {
    const messageMaxChars = 50;
    console.log("displaying available chats")
    let displayLoc = document.getElementById("left-chat-body");
    displayLoc.innerHTML = "";
    sse.getAvailableChats().forEach(chat => {
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
        tileTitleSpan.textContent = formatDateTime(chat.chatStartTime);
        tileTitle.appendChild(tileTitleSpan);

        let tileSubtitle = document.createElement("small");
        tileSubtitle.classList.add("tile-subtitle");
        //skip is messages aren't here yet
        if (chat.messages.length === 0) {
            return;
        }
        // Truncate the message string if it exceeds maxChars
        let truncatedMessage = chat.messages[0].message.slice(0, messageMaxChars);
        if (chat.messages[0].message.length > messageMaxChars) {
            truncatedMessage += "...";
        }
        tileSubtitle.textContent = truncatedMessage;
        let tileAction = document.createElement("div");
        tileAction.classList.add("tile-action");
        tilebutton = document.createElement("button");
        tilebutton.classList.add("btn", "btn-sm");
        tilebutton.textContent = "Join";
        tilebutton.addEventListener("click", function() {
            sockets.connect(chat.chatuuid, myid);
          });
        tileAction.appendChild(tilebutton);

        tileContent.appendChild(tileTitle);
        tileContent.appendChild(tileSubtitle);
        tileContent.appendChild(tileAction);
        tile.appendChild(tileContent);
        displayLoc.appendChild(tile);
    });
}

function displayMyChats() {
    const messageMaxChars = 50;
    let displayLoc = document.getElementById("left-chat-body");
    displayLoc.innerHTML = "";
    sse.getMyChats(myid).forEach(chat => {
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
        tileTitleSpan.textContent = formatDateTime(chat.chatStartTime);
        tileTitle.appendChild(tileTitleSpan);

        let tileSubtitle = document.createElement("small");
        tileSubtitle.classList.add("tile-subtitle");
        
        //skip is messages aren't here yet
        if (chat.messages.length === 0) {
            return;
        }
        // Truncate the message string if it exceeds maxChars
        let truncatedMessage = chat.messages[0].message.slice(0, messageMaxChars);
        if (chat.messages[0].message.length > messageMaxChars) {
            truncatedMessage += "...";
        }
        tileSubtitle.textContent = truncatedMessage;
        //skip is messages aren't here yet
        if (chat.messages.length === 0) {
            return;
        }
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
        tilebutton = document.createElement("button");
        tilebutton.classList.add("btn", "btn-sm");
        tilebutton.textContent = "Join";
        tilebutton.addEventListener("click", function() {
            sockets.connect(chat.chatuuid, myid);
          });
        tileAction.appendChild(tilebutton);
        
        tileContent.appendChild(tileTitle);
        tileContent.appendChild(tileSubtitle);
        tileContent.appendChild(tileAction);
        tile.appendChild(tileContent);
        displayLoc.appendChild(tile);
    });
}

function displayAllChats() {
    const messageMaxChars = 50;
    let displayLoc = document.getElementById("left-chat-body");
    displayLoc.innerHTML = "";
    sse.getAllChats().forEach(chat => {
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
        tileTitleSpan.textContent = formatDateTime(chat.chatStartTime);
        tileTitle.appendChild(tileTitleSpan);

        let tileSubtitle = document.createElement("small");
        tileSubtitle.classList.add("tile-subtitle");
        
        //skip is messages aren't here yet
        if (chat.messages.length === 0) {
            return;
        }
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
        tilebutton = document.createElement("button");
        tilebutton.classList.add("btn", "btn-sm");
        tilebutton.textContent = "Join";
        tilebutton.addEventListener("click", function() {
            sockets.connect(chat.chatuuid, myid);
          });
        tileAction.appendChild(tilebutton);
        
        tileContent.appendChild(tileTitle);
        tileContent.appendChild(tileSubtitle);
        tileContent.appendChild(tileAction);
        tile.appendChild(tileContent);
        displayLoc.appendChild(tile);
    });
}

function formatDateTime(dateTime) {
    const date = new Date(dateTime);
    const options = { hour12: false, hour: 'numeric', minute: 'numeric', second: 'numeric' };
    return date.toLocaleTimeString(undefined, options);
}

function reloadChatMessages() {
    let guid = sockets.getActiveConnection();
    if (guid) {
        sockets.renderMessagesToChatInterface(guid);
    };
}
class SSEvents {
    constructor() {
        this.allchats = [];
        this.eventSource = new EventSource("/chatstream");
        this.eventSource.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.allchats = message;
            console.log(this.allchats)
        };

        this.eventSource.onerror = (error) => {
            console.error("EventSource error:", error);  
            // Reconnect after a delay
            setTimeout(() => {
                this.eventSource.close();
                this.eventSource = new EventSource("/chatstream");
            }, 5000);            
        }
    }

    getAllChats() {
        return this.allchats;
    }

    getMyChats(myid) {
        return this.allchats.filter(chat => hasMeInChat(chat, myid));
    }

    getAvailableChats() {
        return this.allchats.filter(chat => freeForPickup(chat, myid));
    }


    //returns in the format name: message, for use with the socket Connections
    getMessagesFromGuid(guid) {
        let messages = [];
        let chat = this.allchats.find(chat => chat.chatuuid === guid);
        chat.messages.forEach(message => {
            chat.participants.forEach(participant => {
                if (participant.userid === message.userid) {
                    messages.push(`${participant.name}: ${message.message}`);
                }
            });
        });
        return messages;
    }

    getUserInfo(guid, name) {
        let chat = this.allchats.find(chat => chat.chatuuid === guid);
        chat.forEach(participant => {
            if (participant.name === name) {
                return participant;
            }
        });
    }
}

function hasMeInChat(chatArray, myid) {
    for (const participant of chatArray.participants) {
      if (participant.userid === myid && participant.active === true) {
        return true;
      }
    }
  return false;
}

function freeForPickup(chatArray) {
    for (const participant of chatArray.participants) {
      if (participant.internal === true && participant.active === true) {
        return false;
      }
    }
  return true;
}
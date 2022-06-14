var divOfMessagesElement = document.getElementsByClassName("chat-history")[0]
var input = document.getElementById("inputMessage");
var socket = new WebSocket("ws://localhost:8080/ws/");

socket.onmessage = function (e) {
    var listOfMessagesElement = divOfMessagesElement.children[0];
    
    var messages = e.data.split('|');

    var newMessage = document.createElement("li");
    var nMsgChildBody = document.createElement("div");
    var nMsgChildTime = document.createElement("div");
    var nMsgChildTimeChild = document.createElement("span");
    
    newMessage.classList.add("clearfix");
    nMsgChildBody.classList.add("message", "my-message", "float-right");
    nMsgChildTime.classList.add("message-data", "text-right");
    nMsgChildTimeChild.classList.add("message-data-time");

    // if (messages[2] === "Sender"){
    //     newMessage.classList.add("clearfix");
    //     nMsgChildBody.classList.add("message", "my-message", "float-right");
    //     nMsgChildTime.classList.add("message-data", "text-right");
    //     nMsgChildTimeChild.classList.add("message-data-time");
    // } else {
    //     newMessage.classList.add("clearfix");
    //     nMsgChildBody.classList.add("message", "other-message", "float-left");
    //     nMsgChildTime.classList.add("message-data", "text-left");
    //     nMsgChildTimeChild.classList.add("message-data-time");
    // }

    nMsgChildTimeChild.innerHTML += messages[0]
    nMsgChildBody.innerHTML += messages[1]

    nMsgChildTime.appendChild(nMsgChildTimeChild)
    
    newMessage.appendChild(nMsgChildTime);
    newMessage.appendChild(nMsgChildBody);

    listOfMessagesElement.appendChild(newMessage);
};

function send() {
    socket.send(input.value);
    input.value = "";
}

document.getElementById("inputForm").onsubmit = function () {
    if (!socket) {
        return false;
    }

    if (!input.value) {
        return false;
    }
    
    send();

    return false;
}
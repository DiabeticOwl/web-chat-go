var messagesList = document.getElementById("chat-history");
var input = document.getElementById("inputMessage");
var socket = new WebSocket("ws://" + document.location.host + "/ws/");

socket.onmessage = function (e) {
    var message = e.data.split("|");

    var newMessage = document.createElement("div");
    
    newMessage.style.color = message[0];
    newMessage.innerText = message[1];

    messagesList.appendChild(newMessage);

    window.scrollTo(0, document.body.scrollHeight);
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

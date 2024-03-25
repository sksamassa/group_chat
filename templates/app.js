window.addEventListener("DOMContentLoaded", () => {
    const websocket = new WebSocket("ws://" + window.location.host + "/websocket");
    const room = document.getElementById("chat-text");
  
    websocket.addEventListener("message", function (e) {
      const data = JSON.parse(e.data);
      // creating html element
      const p = document.createElement("p");
      p.innerHTML = `<strong>${data.username}</strong>: ${data.text}`;
  
      room.appendChild(p);
      room.scrollTop = room.scrollHeight; // Auto scroll to the bottom
    });
  
    const form = document.getElementById("input-form");
    form.addEventListener("submit", function (event) {
      event.preventDefault();
      let username = document.getElementById("input-username");
      let text = document.getElementById("input-text");
      websocket.send(
        JSON.stringify({
          username: username.value,
          text: text.value,
        })
      );
      text.value = "";
    });
  });
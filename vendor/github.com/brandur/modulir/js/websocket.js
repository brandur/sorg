function connect() {
  var url = "ws://localhost:{{.Port}}/websocket";

  console.log(`Connecting to Modulir: ${url}`);
  var socket = new WebSocket(url);

  socket.onclose = function(event) {
    console.log("Websocket connection closed or unable to connect; starting reconnect timeout");

    // Allow the last socket to be cleaned up.
    socket = null;

    // Set an interval to continue trying to reconnect periodically until we
    // succeed.
    setTimeout(function() {
      connect();
    }, 5000)
  }

  socket.onmessage = function(event) {
    console.log(`Received event of type '${event.type}' data: ${event.data}`);

    var data = JSON.parse(event.data);

    switch(data.type) {
      case "build_complete":
        // 1000 = "Normal closure" and the second parameter is a human-readable
        // reason.
        socket.close(1000, "Reloading page after receiving build_complete");

        console.log("Reloading page after receiving build_complete");
        location.reload(true);

        break;

      default:
        console.log(`Don't know how to handle type '${data.type}'`);
    }
  }

  socket.onopen = function(event) {
    console.log("Websocket connected");
  }
}

connect();

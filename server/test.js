ws = new WebSocket("ws:localhost:10914/battle")

hash = window.location.hash;
var trainer = hash.substring(1, hash.length);

msg1 = {
    "trainer": trainer,
    "Lat": 42.3646,
    "Lng": -71.1028
};

ws.onopen = function(e) {
    ws.send(JSON.stringify(msg1));
};

ws.onmessage = function(e) {
    console.log(e.data);
};

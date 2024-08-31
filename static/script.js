const status = document.querySelector('#status');

// Htmx:wsConnecting
// htmx:wsError

let socket;
let elt;

document.addEventListener('visibilitychange', event => {
	console.log('visibilitychange', document.visibilityState);
	if (socket) {
		socket.send(document.visibilityState, elt);
	}
});

document.body.addEventListener('htmx:wsAfterMessage', event => {
	console.log('message in htmx');
	console.log(event.detail.message);
	console.log('message in htmx end');
});

document.body.addEventListener('htmx:wsOpen', event => {
	console.log('connected');

	socket = event.detail.socketWrapper;
	elt = event.detail.elt;

	status.innerText = 'Connected';
	status.dataset.status = 'connected';
});
document.body.addEventListener('htmx:wsClose', event => {
	console.log('disconnected');
	status.innerText = 'Disconnected';
	status.dataset.status = 'disconnected';
});


		<!--<script>-->
			<!--	let secure = window.location.protocol.includes('https') ? 's':'';-->
				<!--	var socket2 = new WebSocket("ws"+secure+"://" + window.location.host + "/ws");-->
				<!---->
				<!--	socket2.onopen = function(event) {-->
						<!--		console.log("WebSocket connected!");-->
						<!--	}-->
				<!---->
				<!--	//socket2.onmessage = function(event) {-->
						<!--		//	console.log("Received message:", event.data);-->
						<!--		//	document.getElementById("output").innerHTML += event.data + "<br>";-->
						<!--		//}-->
				<!---->
				<!--	function sendMessage() {-->
						<!--		var message = document.getElementById("message").value;-->
						<!--		socket2.send(message);-->
						<!--		document.getElementById("message").value = "";-->
						<!--		console.log("Sent message:", message);-->
						<!--	}-->
				<!--</script>-->

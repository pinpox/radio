const status = document.querySelector('#status');

// Htmx:wsConnecting
// htmx:wsError

let socket;
let elt;

// Document.addEventListener('visibilitychange', event => {
//	console.log('visibilitychange', document.visibilityState);
//	if (socket) {
//		socket.send(document.visibilityState, elt);
//	}
// });
//
document.body.addEventListener('htmx:wsClose', event => {
	console.log('disconnected');
	status.innerText = 'Disconnected';
	status.dataset.status = 'disconnected';
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

const player = document.querySelector('#audio-player');

// Workaround to prevent player from playing when replaced
let audioSource = document.querySelector('#audio-source');
let currentSource = audioSource.src;

const wsContainer = document.querySelector('#ws-container');
wsContainer.addEventListener('htmx:wsAfterMessage', event => {
	audioSource = document.querySelector('#audio-source');
	if (currentSource != audioSource.src) {
		currentSource = audioSource.src;
		player.load();
	}

	// Delete children after n
	const node = document.querySelector('#chatmessages');
	for (const n of node.querySelectorAll('.chatmessage:nth-child(1n+10)')) {
		n.remove();
	}
});

// Volume slider
const volume = document.querySelector('#volume-slider');
volume.addEventListener('change', e => {
	player.volume = e.currentTarget.value;
});

const form = document.querySelector('#form-chat');
const formInput = document.querySelector('#form-chat-input');


    var myEventListener = htmx.on("#form-chat", "submit", function(evt){
		console.log(evt); 

	formInput.value = '';
	});

//u.addEventListener('htmx:wsAfterMessage', event => {
//	audioSource = document.querySelector('#audio-source');
//	if (currentSource != audioSource.src) {
//		currentSource = audioSource.src;
//		player.load();
//	}

//form.addEventListener('submit', e => {
//	console.log('submitted');
//
//	event.preventDefault();
//	formInput.value = '';
//
//	// Player.volume = e.currentTarget.value;
//});

// Pause/Play button
const playPauseButton = document.querySelector('#play-pause-button');
let isPlaying = false;

player.addEventListener('play', () => {
	playPauseButton.textContent = 'pause';
	isPlaying = true;
});

player.addEventListener('pause', () => {
	playPauseButton.textContent = 'play';
	isPlaying = false;
});

playPauseButton.addEventListener('click', () => {
	if (isPlaying) {
		player.pause();
		playPauseButton.textContent = 'Play';
	} else {
		player.play();
		playPauseButton.textContent = 'Pause';
	}
});

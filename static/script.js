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

//document.body.addEventListener('htmx:wsAfterMessage', event => {
//	console.log('message in htmx');
//	console.log(event.detail.message);
//	console.log(event.detail);
//	console.log('message in htmx end');
//});

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

const volume = document.querySelector('#volume-slider');
const player = document.querySelector('#audio-player');
const playPauseButton = document.querySelector('#play-pause-button');
const nextButton = document.querySelector('#button-ws-next');
const previousButton = document.querySelector('#button-ws-prev');
const audioSource = document.querySelector('#audio-source');

// Reload player src after swapping it. This makes sure we don't play two streams at once
// TODO use htmx:afterSwap instead, this fails sometimes
//https://htmx.org/events/#htmx:afterSwap
htmx.on("#button-ws-next", "click", function(evt){ player.load(); });
htmx.on("#button-ws-prev", "click", function(evt){ player.load(); });

// Volume slider
volume.addEventListener('change', e => {
	player.volume = e.currentTarget.value;
});

// Player.setAttribute('src',theNewSource); //change the source
// player.load(); //load the new source
// player.play(); //play

let isPlaying = false;

playPauseButton.addEventListener('click', () => {
	if (isPlaying) {
		player.pause();
		playPauseButton.textContent = 'Play';
	} else {
		player.play();
		playPauseButton.textContent = 'Pause';
	}

	isPlaying = !isPlaying;
});

// Register the service worker so the app is installable as a PWA.
if ('serviceWorker' in navigator) {
	window.addEventListener('load', () => {
		navigator.serviceWorker.register('/sw.js').catch(err => {
			console.warn('Service worker registration failed:', err);
		});
	});
}

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

	// Delete children after n (old messages)
	const node = document.querySelector('#chatmessages');
	for (const n of node.querySelectorAll('.chatmessage:nth-child(1n+10)')) {
		n.remove();
	}
});

// Media Session API: surfaces playback metadata and controls to the OS
// (Android lock screen, Bluetooth car head units, headset buttons, etc.).
const stationNameEl = document.querySelector('#station-name');
const stationTitleEl = document.querySelector('#station-title');

function updateMediaSessionMetadata() {
	if (!('mediaSession' in navigator)) return;

	// #station-name is rendered as "[Name]" inside an <a>; strip the brackets.
	const rawName = (stationNameEl.innerText || '').trim();
	const name = rawName.replace(/^\[/, '').replace(/\]$/, '');
	const title = (stationTitleEl.innerText || '').trim();

	navigator.mediaSession.metadata = new MediaMetadata({
		title: title || name || '0cx Radio',
		artist: name || '0cx Radio',
		album: '0cx Radio',
		artwork: [
			{ src: '/static/icon-192.png', sizes: '192x192', type: 'image/png' },
			{ src: '/static/icon-512.png', sizes: '512x512', type: 'image/png' },
		],
	});
}

if ('mediaSession' in navigator) {
	navigator.mediaSession.setActionHandler('play', () => {
		document.querySelector('#play-pause-button').click();
	});
	navigator.mediaSession.setActionHandler('pause', () => {
		document.querySelector('#play-pause-button').click();
	});
	navigator.mediaSession.setActionHandler('previoustrack', () => {
		document.querySelector('#button-ws-prev').click();
	});
	navigator.mediaSession.setActionHandler('nexttrack', () => {
		document.querySelector('#button-ws-next').click();
	});

	player.addEventListener('play', () => {
		navigator.mediaSession.playbackState = 'playing';
	});
	player.addEventListener('pause', () => {
		navigator.mediaSession.playbackState = 'paused';
	});

	// Refresh metadata whenever the server pushes new station info.
	wsContainer.addEventListener('htmx:wsAfterMessage', updateMediaSessionMetadata);
	updateMediaSessionMetadata();
}

// Volume slider
const volume = document.querySelector('#volume-slider');
volume.addEventListener('change', e => {
	player.volume = e.currentTarget.value;
});

const form = document.querySelector('#form-chat');
const formInput = document.querySelector('#form-chat-input');

htmx.on('#form-chat', 'submit', event => { formInput.value = ''; });

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

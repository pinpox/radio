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
const mediaArtwork = [
	{ src: new URL('/static/icon-192.png', location.href).href, sizes: '192x192', type: 'image/png' },
	{ src: new URL('/static/icon-512.png', location.href).href, sizes: '512x512', type: 'image/png' },
];

function updateMediaSessionMetadata() {
	if (!('mediaSession' in navigator)) return;

	// Re-query each call: htmx hx-swap-oob *replaces* the elements with
	// matching IDs, so any cached references would point at detached nodes
	// and return empty text after the first WS update.
	const stationNameEl = document.querySelector('#station-name');
	const stationTitleEl = document.querySelector('#station-title');
	if (!stationNameEl || !stationTitleEl) return;

	// #station-name is rendered as "[Name]" inside an <a>; strip the brackets.
	const rawName = (stationNameEl.innerText || '').trim();
	const name = rawName.replace(/^\[/, '').replace(/\]$/, '');
	const rawTitle = (stationTitleEl.innerText || '').trim();

	const title = rawTitle || name || '0cx Radio';
	const artist = name || '0cx Radio';

	if (!navigator.mediaSession.metadata) {
		navigator.mediaSession.metadata = new MediaMetadata({
			title,
			artist,
			album: '0cx Radio',
			artwork: mediaArtwork,
		});
		return;
	}

	const md = navigator.mediaSession.metadata;
	if (md.title !== title) md.title = title;
	if (md.artist !== artist) md.artist = artist;
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
		updateMediaSessionMetadata();
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

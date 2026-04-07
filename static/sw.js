// Minimal service worker: required so Chrome treats the site as an
// installable PWA. The audio stream is live, so there is nothing useful
// to cache; we just pass every request straight through to the network.

self.addEventListener('install', event => {
	self.skipWaiting();
});

self.addEventListener('activate', event => {
	event.waitUntil(self.clients.claim());
});

self.addEventListener('fetch', event => {
	// No-op: let the browser handle the request normally.
});

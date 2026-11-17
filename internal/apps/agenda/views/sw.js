const CACHE_NAME = 'agenda-nobre-v4';
const ASSETS = [
  '/',
  '/home.js',
  '/style.css',
  '/manifest.json',
  '/sw.js',
  '/assets/logo.png',
  '/agenda/index.html',
  '/agenda/style.css',
  '/agenda/script.js',
  '/agenda/assets/logo.png',
  '/financeiro/index.html',
  '/financeiro/style.css',
  '/financeiro/app.js',
  '/financeiro/assets/logo.png'
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(ASSETS);
    })
  );
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheName !== CACHE_NAME) {
            return caches.delete(cacheName);
          }
        })
      );
    })
  );
});

self.addEventListener('fetch', (event) => {
  if (event.request.mode === 'navigate') {
    event.respondWith(
      fetch(event.request).catch(() => {
        return caches.match(event.request);
      })
    );
    return;
  }

  event.respondWith(
    caches.match(event.request).then((response) => {
      return response || fetch(event.request);
    })
  );
});

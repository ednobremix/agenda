const CACHE_NAME = 'portalnobre-nobre-v13';
const ASSETS = [
  '/',
  '/html/home.html',
  '/js/home.js',
  '/css/style.css',
  '/manifest.json',
  '/sw.js',
  '/assets/logo.png',
  '/agenda/html/index.html',
  '/agenda/html/login.html',
  '/agenda/html/admin.html',
  '/agenda/html/ajuda.html',
  '/agenda/css/style.css',
  '/agenda/js/script.js',
  '/agenda/js/login.js',
  '/agenda/js/admin.js',
  '/agenda/assets/logo.png',
  '/financeiro/html/index.html',
  '/financeiro/html/login.html',
  '/financeiro/html/admin.html',
  '/financeiro/html/ajuda.html',
  '/financeiro/css/style.css',
  '/financeiro/css/admin.css',
  '/financeiro/css/ajuda.css',
  '/financeiro/js/app.js',
  '/financeiro/js/login.js',
  '/financeiro/js/admin.js',
  '/financeiro/assets/logo.png',
  '/estoque/html/index.html',
  '/estoque/html/login.html',
  '/estoque/html/ajuda.html',
  '/estoque/css/style.css',
  '/estoque/css/ajuda.css',
  '/estoque/js/app.js',
  '/estoque/js/login.js',
  '/estoque/assets/logo.png',
  '/praca/html/index.html',
  '/praca/html/slides.html',
  '/praca/html/galeria3d.html',
  '/praca/html/ajuda.html',
  '/praca/css/style.css',
  '/praca/css/galeria3d.css',
  '/praca/css/ajuda.css',
  '/praca/js/slides.js',
  '/praca/js/galeria3d.js'
];

self.addEventListener('install', (event) => {
  self.skipWaiting();
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
    }).then(() => self.clients.claim())
  );
});

self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);

  if (event.request.method !== 'GET' || url.pathname.includes('/api/')) {
    return;
  }

  const isCodeOrDoc =
    event.request.mode === 'navigate' ||
    url.pathname.endsWith('.js') ||
    url.pathname.endsWith('.css') ||
    url.pathname.endsWith('.html');

  if (isCodeOrDoc) {
    event.respondWith(
      fetch(event.request)
        .then((response) => {
          if (response && response.ok && url.origin === self.location.origin) {
            const copy = response.clone();
            caches.open(CACHE_NAME).then((cache) => cache.put(event.request, copy));
          }
          return response;
        })
        .catch(() => caches.match(event.request))
    );
    return;
  }

  event.respondWith(
    caches.match(event.request).then((response) => {
      return response || fetch(event.request);
    })
  );
});

// Tim-Agent 轻量自愈 Service Worker
// 核心设计原则: 极简防陈旧缓存，网络优先(Network-First)，即时更新接管，确保多端版本实时一致

const CACHE_VERSION = 'v1.0.5';
const CACHE_NAME = 'tim-agent-cache-' + CACHE_VERSION;

// 仅预缓存必须的终端基础库文件
const PRECACHE_ASSETS = [
  '/css/xterm.css',
  '/css/terminal.css',
  '/js/xterm.js',
  '/js/addon-fit.js',
  '/img/icon.png',
  '/manifest.json'
];

// 安装阶段：预缓存核心静态文件，并立即跳过等待
self.addEventListener('install', event => {
  self.skipWaiting();
  event.waitUntil(
    caches.open(CACHE_NAME).then(cache => {
      return cache.addAll(PRECACHE_ASSETS).catch(err => {
        console.warn('[SW] 预缓存部分文件失败 (非致命):', err);
      });
    })
  );
});

// 激活阶段：清理所有过期的旧版本 Cache，并立即接管当前客户端
self.addEventListener('activate', event => {
  event.waitUntil(
    Promise.all([
      // 1. 删除除当前 CACHE_NAME 之外的所有老旧缓存，避免过重缓存
      caches.keys().then(keys => {
        return Promise.all(
          keys.map(key => {
            if (key !== CACHE_NAME) {
              console.log('[SW] 清理过时旧版本缓存:', key);
              return caches.delete(key);
            }
          })
        );
      }),
      // 2. 立即接管所有打开的客户端页面
      self.clients.claim()
    ])
  );
});

// 请求拦截阶段：
// - WebSocket 与 API 接口: 100% 网络直通，绝不拦截缓存
// - HTML 页面与业务 JS/CSS: 网络优先 (Network-First)，仅当完全离线时降级回退到缓存
self.addEventListener('fetch', event => {
  const req = event.request;
  const url = new URL(req.url);

  // 1. WebSocket 请求直接忽略，走原生通道
  if (url.protocol === 'ws:' || url.protocol === 'wss:' || url.pathname.startsWith('/ws/')) {
    return;
  }

  // 2. API 数据接口与登录接口：Network-Only，绝不进行本地缓存
  if (url.pathname.startsWith('/api/') || req.method !== 'GET') {
    return;
  }

  // 3. 页面导航 (如 /terminal, /, /login) 与核心静态资源:
  // 严格采用 Network-First (网络优先)，获取成功时更新本地副本，离线断网时降级到缓存
  event.respondWith(
    fetch(req)
      .then(networkResponse => {
        // 请求成功 (包含 200)，将最新内容克隆写入缓存
        if (networkResponse && networkResponse.status === 200) {
          const resClone = networkResponse.clone();
          caches.open(CACHE_NAME).then(cache => {
            cache.put(req, resClone).catch(() => {});
          });
        }
        return networkResponse;
      })
      .catch(async () => {
        // 网络请求失败 (离线/断网时)，尝试从 Cache 回退
        const cachedResponse = await caches.match(req);
        if (cachedResponse) {
          return cachedResponse;
        }
        // 如果是 HTML 导航且缓存未命中，尝试回退到已缓存的 /terminal
        if (req.mode === 'navigate') {
          return caches.match('/terminal');
        }
        return new Response('Network Offline', { status: 503, statusText: 'Service Unavailable' });
      })
  );
});

// 监听宿主页面主动发来的指令 (如收到新版本通知后强制跳过等待)
self.addEventListener('message', event => {
  if (event.data && event.data.action === 'skipWaiting') {
    self.skipWaiting();
  }
});

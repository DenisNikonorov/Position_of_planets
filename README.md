# Лабораторная работа: адаптивность, PWA и развертывание

React + Redux Toolkit (фильтры, авторизация, данные), кодогенерация API из
swagger, адаптивная вёрстка, PWA (manifest + service worker), развертывание
статики на GitHub Pages (mock) и HTTPS-реверс-прокси для доступа с телефона
по локальной сети.

## Запуск

```bash
# 1. Бэкенд (из каталога backend): Postgres/Redis/MinIO через docker compose,
#    миграция и сервер на :8080 (swagger: http://localhost:8080/swagger/doc.json)
go run ./cmd/migrate
go run ./cmd/planets.go

# 2. Фронтенд (из каталога frontend): dev-сервер на :3000
npm install
npm run dev

# 3. Кодогенерация API-клиента (бэкенд должен быть запущен)
npm run generate-api
```

Lint и сборка: `npm run lint`, `npm run build` (проверка типов + сборка).

## Адаптивность

Основные правила — в `src/planets.css`:

| Элемент | Значения |
| --- | --- |
| Контейнер (`body`, `.bottom-nav`) | `max-width: 1280px`, центрирование |
| `.tiles-grid` | 1 колонка; `@media (min-width: 700px)` — 2; `@media (min-width: 1040px)` — 3; `gap: 12px` |
| `.tile-card` | `min-width: 300px` (на экранах `<360px` — `0`), ширина колонки `minmax(min(300px, 100%), 1fr)` |
| Лента | `.feed-container { height: calc(100dvh - 64px) }`, на `≥1000px` описание `width: min(560px, calc(100% - 96px))` |
| Добавление/вход | `.add-container { max-width: 700px; margin: 0 auto }` |
| Формы | `.form-row` на `≤460px` разворачивается в колонку |

Ширина карточки в плитке зависит от ширины экрана (колонки `1fr`), но не
меньше 300px; количество колонок — 1 / 2 / 3 по брейкпоинтам выше.

## PWA

Настроено через `vite-plugin-pwa` (`vite.config.ts`):

- `manifest`: имя «Положение планет», `short_name` «Планеты», `display: standalone`,
  `orientation: portrait-primary`, фон/тема `#0a0a0a`, `lang: ru`,
  иконки `logo192.png`, `logo512.png`, `maskable-512.png`;
- service worker генерируется (`registerType: 'autoUpdate'`), регистрируется в
  `src/main.tsx` через `registerSW()`;
- в precache попадают **дефолтные фото и видео** (`media/default_photo.png`,
  `media/default_video.mp4`) — `workbox.globPatterns` c `mp4` и лимит
  `maximumFileSizeToCacheInBytes: 8 МБ`, поэтому офлайн-режим и mock на Pages
  показывают медиа;
- `devOptions.enabled: true` — SW работает и в dev-режиме.

Проверка: DevTools → Application → Manifest (без ошибок), Service Workers
(active), Cache Storage (файлы из precache). На телефоне — «Добавить на главный
экран» (нужен HTTPS).

## GitHub Pages (только mock)

```bash
npm run deploy   # = build:pages + gh-pages -d dist
```

- `build:pages` собирает с `GITHUB_PAGES=1 VITE_MOCK=1`: `base` =
  `/Position_of_planets/`, а thunk'и при `VITE_MOCK=1` сразу возвращают mock
  (`PLANETS_MOCK`) без обращения к бэкенду;
- роутер получает `basename: import.meta.env.BASE_URL`;
- ссылки в navbar — `Link` (basename подставляется автоматически);
- дефолтные медиа строятся от `import.meta.env.BASE_URL`, поэтому совпадают
  с ключами precache;
- `public/404.html` возвращает SPA-маршрут при перезагрузке внутренней страницы.

После первого деплоя включить Pages: **Settings → Pages → Source: branch
`gh-pages`, folder `/ (root)`**. Адрес:
`https://<username>.github.io/Position_of_planets/`.

## HTTPS (локальная сеть)

Node.js-сервер Vite — реверс-прокси (`/api` → :8080, `/minio` → :9000) — по
умолчанию работает по HTTPS с самоподписанным сертификатом:

```bash
cd frontend
openssl req -x509 -newkey rsa:2048 -sha256 -days 365 -nodes \
  -keyout certs/cert.key -out certs/cert.crt \
  -subj "/CN=Position_of_planets" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1,IP:<IP-ноутбука>"
npm run dev:https   # HTTPS=1 vite --host
```

Сертификаты лежат в `frontend/certs/` и **не коммитятся** (`.gitignore`).
`vite.config.ts` при `HTTPS=1` подключает `server.https`, `host: true`,
`allowedHosts: true`.

Телефон и ноутбук — в одной сети; открыть `https://<IP-ноутбука>:3000`,
принять предупреждение самоподписанного сертификата. Запросы к бэкенду и MinIO
идут через тот же HTTPS-реверс-прокси, поэтому PWA и service worker работают
(secure context). IP видно командой `ip -4 addr`; при смене IP сертификат нужно
перегенерировать с новым `IP:`.

### WSL2

Во WSL2 адрес `eth0` (например `172.21.x.x`) — внутренняя NAT-сеть, телефон до
неё не достучится. Нужно пробросить порт с Windows (PowerShell от админа):

```powershell
$wsl = (wsl hostname -I).Trim().Split(' ')[0]
netsh interface portproxy add v4tov4 listenaddress=0.0.0.0 listenport=3000 connectaddress=$wsl connectport=3000
New-NetFirewallRule -DisplayName "WSL Planets 3000" -Direction Inbound -Protocol TCP -LocalPort 3000 -Action Allow
```

Телефон открывает `https://<IP-Windows>:3000` (сертификат должен содержать этот
IP в SAN); бэкенд/MinIO наружу не пробрасываются — их проксирует Vite.
IP WSL меняется после `wsl --shutdown` — правило portproxy нужно пересоздать.

## Проксирование (CORS не нужен)

Маршруты в `frontend/vite.config.ts`:

| Маршрут | Цель | Назначение |
| --- | --- | --- |
| `/api` | `http://localhost:8080` | REST-бэкенд (спецификация — `/swagger/doc.json`) |
| `/minio` | `http://localhost:9000` (префикс `/minio` срезается) | объекты MinIO: `image_url`/`url_video` вида `/minio/planets/<file>` |

В `backend/config/config.toml` задан `MinioPublicURL = "/minio"`, поэтому
ссылки на медиа относительные и идут через прокси.

## Redux (Redux Toolkit)

`src/store/index.ts` — `configureStore` c редьюсерами `planets` и `user`,
Redux DevTools включены по умолчанию.

### Слайс `planets` (`src/store/planetsSlice.ts`)

| Часть | Описание |
| --- | --- |
| `maxDist` | значение слайдера фильтра (хранится в Redux) |
| `appliedMaxDist` | применённый фильтр; `null` — показать все планеты |
| `planets`, `feed`, `draft` | список, текущая планета ленты, черновик |
| `loading` | во время запросов включает анимацию `.page-loading` |
| `error` | текст ошибки |

Thunks (`createAsyncThunk`, вызывают сгенерированный `src/api/Api.ts`):
`getPlanetsList`, `getFeed`, `createPlanet`, `getDraft`, `publishPlanet`,
`likePlanet`, `deletePlanet`. При `VITE_MOCK=1` (Pages) `getPlanetsList`/
`getFeed` возвращают mock; без флага mock используется как fallback, если
бэкенд недоступен **и пользователь — гость** (с `console.warn`).

### Слайс `user` (`src/store/userSlice.ts`)

`{ login, isAuthenticated }` + plain-действия `login`/`logout` (без thunk).
Вход/выход выполняются **ручным axios** вне Redux:

- `POST /api/users/auth` → токен сохраняется в **localStorage `jwt`**
  (`src/lib/token.ts`) и в `setAuthToken(token)` (интерцептор axios подставляет
  `Bearer` во все запросы, включая публичные list/feed — по ним бэкенд считает
  `liked`/`is_owner`);
- `POST /api/users/deauth` → localStorage очищается, `logout`, **сброс фильтра**
  и обновление списка/ленты уже как гость.

После F5 Redux сбрасывается в состояние гостя, а JWT остаётся в localStorage
(Application → Local Storage).

## Страницы

| Маршрут | Страница | Описание |
| --- | --- | --- |
| `/` | Лента | `GET /api/planets/feed?id=&next=true`, лайк (`liked` 0/1), клавиши ↑/↓, видео с fallback на дефолтное |
| `/tile` | Плитка | фильтр «Макс. расстояние» из Redux → `GET /api/planets?maxDist=`; карточка: реальное фото → видео → дефолт; лайк; «Удалить» при `is_owner=1` |
| `/add` | Добавление | шаг 1 — dropzone + название (`POST /api/planets`, черновик); шаг 2 — описание/расстояние/дата (`PUT /api/planets/:id`); только авторизованному |
| `/login` | Вход | логин + пароль, токен в localStorage |

Нижняя навигация (`PlanetsNavbar`): Лента · Добавить · Плитка · Вход/Выход.
У гостя «Добавить» неактивно, лайки/удаление видимы, но `disabled`.

## Структура

```
frontend/
├── vite.config.ts              # base (Pages), PWA, HTTPS, прокси
├── certs/                      # локальные сертификаты (gitignored)
├── public/
│   ├── favicon.svg             # иконка приложения
│   ├── logo192.png, logo512.png, maskable-512.png
│   ├── 404.html                # SPA-фолбэк для GitHub Pages
│   └── media/                  # дефолтные фото и видео (в precache)
└── src/
    ├── main.tsx                # Provider + роутер (basename) + registerSW
    ├── hooks.ts                # useAppDispatch / useAppSelector
    ├── api/
    │   ├── Api.ts              # сгенерированный клиент (swagger-typescript-api, axios)
    │   └── index.ts            # instance: baseURL "/api" + Bearer-интерцептор
    ├── lib/token.ts            # localStorage jwt: get/set/clear
    ├── store/
    │   ├── index.ts            # configureStore (devtools включены)
    │   ├── planetsSlice.ts     # фильтр + данные + thunks + mock
    │   └── userSlice.ts        # login/logout (plain actions)
    ├── modules/
    │   ├── planets.ts          # Planet, выбор медиа (пути от BASE_URL)
    │   └── planetsMock.ts      # mock для Pages/гостя
    ├── components/             # PlanetsNavbar/Card/Filter/Loading
    └── pages/                  # Feed/Tile/Add/Login
```

## Контрольные вопросы

**Flux** — архитектура с однонаправленным потоком данных:
`Action → Dispatcher → Store → View`; View порождает новые Action, состояние
меняется только через dispatcher.

**Redux** — реализация Flux: `store` хранит единственное состояние; `action` —
объект `{ type, payload }`; `reducer` — чистая функция `(state, action) => state`;
`dispatch(action)` отправляет действие в store; компоненты подписываются на
изменения (`useSelector`). Асинхронность — через middleware (`redux-thunk`).

**Виды нативных приложений**: PWA (веб + manifest/service worker, установка на
экран, офлайн), Tauri (Rust + системный webview, малый размер), Electron
(Chromium + Node.js, тяжёлый), Capacitor/Cordova (webview-обёртка), React Native
и Flutter (нативные UI), .NET MAUI.

**GitHub Pages** — бесплатный статический хостинг из GitHub-репозитория (ветка
`gh-pages`), только статика без бэкенда, доступен по
`https://<user>.github.io/<repo>/`; поэтому деплой работает с mock.

## Кодогенерация API

`npm run generate-api` — `swagger-typescript-api` читает живую спецификацию
`http://localhost:8080/swagger/doc.json` и создаёт `src/api/Api.ts`
(httpClient — axios). Клиент используется в thunk-ах; авторизация/деавторизация
сознательно выполняются ручным axios, без кодогенерации и без thunk.

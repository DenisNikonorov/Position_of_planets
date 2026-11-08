# Лабораторная работа №6

React + Redux Toolkit: управление состоянием (фильтры, авторизация, данные),
axios и кодогенерация API из swagger, авторизация с токеном в localStorage, проксирование
запросов к бэкенду и MinIO.

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

## Проксирование (CORS не нужен)

Маршруты в `frontend/vite.config.ts`:

| Маршрут | Цель | Назначение |
| --- | --- | --- |
| `/api` | `http://localhost:8080` | REST-бэкенд (спецификация — `/swagger/doc.json`) |
| `/minio` | `http://localhost:9000` (префикс `/minio` срезается) | объекты Silo/MinIO: `image_url`/`url_video` вида `/minio/planets/<file>` |

В `backend/config/config.toml` задан `MinioPublicURL = "/minio"`, поэтому
ссылки на медиа относительные и идут через прокси (в Network видны адреса
объектов MinIO).

## Redux (Redux Toolkit)

`src/store/index.ts` — `configureStore` c редьюсерами `planets` и `user`,
Redux DevTools включены по умолчанию.

### Слайс `planets` (`src/store/planetsSlice.ts`)

| Часть | Описание |
| --- | --- |
| `maxDist` | значение слайдера фильтра (хранится в Redux) |
| `appliedMaxDist` | применённый фильтр; `null` — показать все планеты |
| `planets`, `feed`, `draft` | список, текущая планета ленты, черновик |
| `loading` | во время запросов включает анимацию `.page-loading` (затемнение + спиннер блокируют кнопки) |
| `error` | текст ошибки |

Thunks (`createAsyncThunk`, вызывают сгенерированный `src/api/Api.ts`):
`getPlanetsList`, `getFeed`, `createPlanet`, `getDraft`, `publishPlanet`,
`likePlanet`, `deletePlanet`. Если бэкенд недоступен **и пользователь — гость**,
`getPlanetsList`/`getFeed` возвращают mock-коллекцию из `planetsMock.ts`
(с `console.warn`), авторизованному показывается ошибка.

### Слайс `user` (`src/store/userSlice.ts`)

`{ login, isAuthenticated }` + plain-действия `login`/`logout` (без thunk).
Вход/выход выполняются **ручным axios** вне Redux:

- `POST /api/users/auth` → токен сохраняется в **localStorage `jwt`** (`src/lib/token.ts`)
  и в `setAuthToken(token)` (интерцептор axios подставляет `Bearer` во все запросы,
  включая публичные list/feed — по ним бэкенд считает `liked`/`is_owner`);
- `POST /api/users/deauth` → localStorage очищается, `logout`, **сброс фильтра**
  (`resetFilter`) и обновление списка/ленты уже как гость.

После F5 Redux сбрасывается в состояние гостя (заголовок `Authorization`
не отправляется), а JWT остаётся в localStorage — его видно в Application
→ Local Storage и можно использовать вручную (Insomnia/Postman).

## Страницы

| Маршрут | Страница | Описание |
| --- | --- | --- |
| `/` | Лента | `GET /api/planets/feed?id=&next=true`, лайк (`liked` 0/1, сердечко `.liked`), клавиши ↑/↓ (вверх — по локальной истории), видео планеты с fallback на `/media/default_video.mp4` |
| `/tile` | Плитка | фильтр «Макс. расстояние» из Redux → `GET /api/planets?maxDist=`; карточка: медиа (реальное фото → реальное видео → дефолт), лайк (клик только авторизованного), **«Удалить»** (`.filter-btn`, как «Найти») только при `is_owner=1` |
| `/add` | Добавление | шаг 1 — шаблон `telescope.html` (dropzone + название → `POST /api/planets` multipart, черновик); шаг 2 — шаблон `planet_publish.tmpl` 1:1 (описание/расстояние/дата → `PUT /api/planets/:id`, `perihelion` = расстоянию, т.к. в шаблоне поля нет, а API его требует); доступен только авторизованному |
| `/login` | Вход | ровно 2 поля (логин, пароль) + кнопка «Войти», стили текущего дизайна (`.form-group`, `.form-input`, `.step-btn`), без лишнего текста |

Нижняя навигация (`PlanetsNavbar`): Лента · Добавить · Плитка · Вход/Выход.
«Добавить» у гостя — `.disabled`; после входа пункт «Вход» заменяется на
«Выход» (иконка выхода), клик выполняет deauth и сброс фильтра.

Лайки и удаление у гостя: кнопки видимы, но неактивны (`disabled`).

## Структура

```
src/
├── main.tsx                    # Provider + роутер (/, /tile, /add, /login)
├── hooks.ts                    # useAppDispatch / useAppSelector
├── api/
│   ├── Api.ts                  # сгенерированный клиент (swagger-typescript-api, axios)
│   └── index.ts                # instance: baseURL "/api" + Bearer из securityData
├── lib/token.ts                # localStorage jwt: get/set/clear
├── store/
│   ├── index.ts                # configureStore (devtools включены)
│   ├── planetsSlice.ts         # фильтр + данные + createAsyncThunk + mock fallback
│   └── userSlice.ts            # login/logout (plain actions)
├── modules/
│   ├── planets.ts              # Planet = Required<DsPlanetSerializer>, выбор медиа
│   └── planetsMock.ts          # mock в форме PlanetSerializer (для гостя без бэка)
├── components/
│   ├── PlanetsNavbar.tsx       # Вход/Выход, deauth, сброс фильтра
│   ├── PlanetsCard.tsx         # карточка: лайк 0/1, удаление (is_owner)
│   ├── PlanetsFilter.tsx       # фильтр из Redux (setMaxDist / getPlanetsList)
│   └── PlanetsLoading.tsx      # оверлей-спиннер, блокирующий кнопки
└── pages/
    ├── PlanetsFeedPage.tsx     # лента (thunk feed, лайк, история ↑/↓)
    ├── PlanetsTilePage.tsx     # плитка (thunk списка, фильтр из Redux)
    ├── PlanetsAddPage.tsx      # шаг 1 (telescope.html) + шаг 2 (planet_publish.tmpl)
    └── PlanetsLoginPage.tsx    # ручной axios → localStorage + dispatch(login)
```

## Кодогенерация API

`npm run generate-api` — `swagger-typescript-api` читает живую спецификацию
`http://localhost:8080/swagger/doc.json` и создаёт `src/api/Api.ts`
(httpClient — axios). Клиент используется в thunk-ах; авторизация/деавторизация
сознательно выполняются ручным axios, без кодогенерации и без thunk.

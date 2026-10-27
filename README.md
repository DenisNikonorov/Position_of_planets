# Обсерватория — фронтенд (React + TypeScript)

Базовое SPA для гостя: просмотр планет (услуг обсерватории) в двух видах — **лента** и **плитка**.
Визуальное оформление полностью повторяет исходный дизайн проекта (тёмная тема `#0a0a0a`,
акцент `#e2b340`, ширина колонки 430px) — CSS из корневого `style.css` скопирован в `src/planets.css`
и не изменялся (добавлено только правило `.nav-item.disabled` для неактивных пунктов меню).

Данные берутся **только из mock-коллекции** (`src/modules/planetsMock.ts`), бэкенд не используется.
Карточки и лента используют **изображение и видео по умолчанию** из `public/media/`
(`/media/default_photo.png`, `/media/default_video.mp4`), если соответствующее поле планеты пустое.

## Запуск

```shell
npm install
npm run dev   # http://localhost:3000
npm run build # production-сборка
```

## Страницы

### 1. Лента `/` (`PlanetsFeedPage`)

Повторяет `templates/planets.html`: вертикальный фид с видео планеты на весь экран.

| Элемент | Класс (исходный дизайн) | Описание |
| --- | --- | --- |
| Видео на весь экран | `.feed-media` | `urlVideo` планеты; если поле пустое → `/media/default_video.mp4` |
| Кнопка лайка | `.like-btn` (+`.liked`) | SVG-сердце, ставит/снимает лайк (состояние `liked`), счётчик `.feed-likes-count` под кнопкой |
| Статистика | `.feed-stats`, `.feed-date` | Расстояние от Солнца в AU и дата наблюдения (иконка календаря) |
| Кнопка «Далее» | `.next-btn` | Золотой круг с шевроном: переход к следующей планете (по кругу) |
| Описание | `.feed-description`, `.planet-name`, `.desc-text` | Название и описание; раскрытие «больше/меньше» через чекбокс `.desc-checkbox` (чистый CSS, как в оригинале) |

Хуки: `useState` (индекс планеты, лайки), `useEffect` — клавиши ↑/↓ переключают планету
(подписка на событие + очистка при размонтировании).

### 2. Плитка `/tile` (`PlanetsTilePage`)

Повторяет `templates/observatory.html`: фильтр `.filter-block` + сетка `.tiles-grid` (2 колонки).

Фильтр (компонент `PlanetsFilter`, кнопка — react-bootstrap `Button` с классом оригинала `.filter-btn`):

| Поле | Оригинальный класс | Описание |
| --- | --- | --- |
| «Макс. расстояние (AU)» | `.filter-slider` + `.filter-value` | Слайдер 0–12 AU (аналог «цены» услуги), значение подписано в `.filter-value`, метки 0/6/12 в `.filter-marks` |
| «Найти» | `.filter-btn` | Применяет фильтр (как submit-формы в оригинале) |

Карточка (`PlanetsCard`, классы `.tile-card`/`.tile-card-info`/`.tile-card-meta`):
медиа (видео, если `hasVideo`, иначе изображение; fallback — `/media/default_photo.png`
или `/media/default_video.mp4`), ниже — название `.tile-card-name`, дата `.tile-card-date`,
лайки `.tile-card-likes`, расстояние `.tile-card-distance`.

### 3. Нижняя навигация (`PlanetsNavbar`)

Исходный `.bottom-nav` из `style.css`, 4 пункта `.nav-item` с оригинальными SVG-иконками:

| Пункт | Доступен гостю | Переход |
| --- | --- | --- |
| Лента | да | `/` (подсветка `.active` по URL) |
| Добавить | нет | пункт `.disabled` (приглушён, `pointer-events: none`) |
| Плитка | да | `/tile` (подсветка `.active` по URL) |
| Вход | нет | пункт `.disabled` |

Панель закреплена внизу (`position: fixed`), активный пункт — золотой `#e2b340`.

## Структура

```
src/
├── main.tsx                    # подключение роутера, bootstrap.css и planets.css
├── App.tsx                     # оболочка: Outlet + нижняя навигация
├── planets.css                 # исходный дизайн проекта (copy корневого style.css)
├── modules/
│   ├── planets.ts              # интерфейс Planet
│   └── planetsMock.ts          # mock-коллекция планет
├── components/
│   ├── PlanetsNavbar.tsx       # нижняя навигация (4 пункта)
│   ├── PlanetsCard.tsx         # карточка плитки (+ дефолтные медиа)
│   └── PlanetsFilter.tsx       # фильтр плитки
└── pages/
    ├── PlanetsFeedPage.tsx     # страница «Лента»
    └── PlanetsTilePage.tsx     # страница «Плитка»
```

## Используемые хуки

- `useState` — индекс текущей планеты в ленте, лайки, значение слайдера и применённое значение фильтра;
- `useEffect` — подписка на клавиши ↑/↓ в ленте (монтирование/размонтирование);
- `useMemo` — вычисление отфильтрованного списка планет на плитке.

Redux и React Context не используются.

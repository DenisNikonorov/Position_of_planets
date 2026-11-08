-- Пользователи (таблица users)
INSERT INTO users (id, login, password, is_moderator)
VALUES
  (1, 'user', 'user', false),
  (2, 'admin', 'admin', true);

-- Планеты (таблица planets)
INSERT INTO planets (id, user_id, image_url, url_video, name, description, distance_from_sun, perihelion, status, observation_date)
VALUES
  (
    1,
    1,
    NULL,
    'Mercury.mp4',
    'Меркурий',
    'Ближайшая к Солнцу планета. Поверхность покрыта кратерами, атмосфера практически отсутствует. Температура колеблется от -180°C ночью до +430°C днём.',
    0.39,
    46.0,
    'published',
    '2026-03-15'
  ),
  (
    2,
    1,
    'Earth.png',
    NULL,
    'Земля',
    'Единственная известная планета с жидкой водой и жизнью. Атмосфера состоит преимущественно из азота и кислорода. Имеет один естественный спутник — Луну.',
    1.00,
    102.9,
    'published',
    '2026-04-02'
  ),
  (
    3,
    1,
    NULL,
    'Mars_spin.mp4',
    'Марс',
    'Красная планета с разреженной атмосферой. Имеет два небольших спутника — Фобос и Деймос. Ржавый цвет поверхности обусловлен оксидом железа.',
    1.52,
    336.1,
    'published',
    '2026-05-20'
  ),
  (
    4,
    1,
    NULL,
    'Jupiter_spin.mp4',
    'Юпитер',
    'Крупнейшая планета Солнечной системы. Газовый гигант с мощной системой вихрей, включая Большое Красное Пятно — шторм, длящийся сотни лет.',
    5.20,
    14.7,
    'published',
    '2026-06-11'
  ),
  (
    5,
    1,
    'Saturn.png',
    NULL,
    'Сатурн',
    'Планета с впечатляющей системой колец из льда и камня. Плотность Сатурна меньше воды — он бы всплыл в достаточно большом океане.',
    9.58,
    95.5,
    'published',
    '2026-07-08'
  );

-- Лайки (таблица likes) - связь многие-ко-многим между пользователями и планетами
INSERT INTO likes (id, user_id, planet_id)
VALUES
  (1, 1, 2),
  (2, 1, 3),
  (3, 1, 4),
  (4, 2, 5);

-- Синхронизация последовательностей первичных ключей с текущими данными
SELECT setval(pg_get_serial_sequence('users','id'), (SELECT COALESCE(MAX(id),1) FROM users));
SELECT setval(pg_get_serial_sequence('planets','id'), (SELECT COALESCE(MAX(id),1) FROM planets));
SELECT setval(pg_get_serial_sequence('likes','id'), (SELECT COALESCE(MAX(id),1) FROM likes));
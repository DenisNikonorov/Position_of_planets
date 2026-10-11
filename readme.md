<h1>Лабораторная работа №3</h1>
<b>Цель</b> работы -  создание веб-сервиса в бэкенде нашей системы для использования его в SPA

<b>Задание</b> - создание веб-сервиса со всей итоговой бизнес логикой (кроме авторизации), подключение его к БД и тестирование в postman

<h2>Структура базы данных</h2>

<b>`users`</b> - пользователи
`id` - `bigint`, идентификатор пользователя, первичный ключ - `id`
`login` - `varchar(25)`, логин, уникальное значение - `login`
`password` - `varchar(100)`, пароль - не отдаётся в JSON
`is_moderator` - `boolean`, признак модератора, по умолчанию `false` - не отдаётся в JSON

<b>`planets`</b> - услуги
`id` - `bigint`, идентификатор планеты, первичный ключ - `id`
`name` - `varchar(50)`, название планеты, обязательное - `name`
`description` - `varchar(255)`, описание планеты - `description`
`distance_from_sun` - `numeric`, расстояние до Солнца, обязательное - `distance_from_sun`
`perihelion` - `numeric`, перигелий - `perihelion`
`observation_date` - `varchar(10)`, дата наблюдения - `observation_date`
`image_url` - `varchar(150)`, имя файла изображения в MinIO - `image_url`
`url_video` - `varchar(150)`, имя файла видео в MinIO - `url_video`
`status` - `varchar(20)`, состояние записи: `draft`, `published`, `deleted` - не отдаётся в JSON
`user_id` - `bigint`, владелец записи, внешний ключ на `users` - не отдаётся в JSON
`created_at` - `timestamp with time zone`, дата и время создания - `created_at`
`updated_at` - `timestamp with time zone`, дата и время изменения - `updated_at`
`likes_count` - `integer`, количество лайков, вычисляется из таблицы `likes` - `likes_count`
`liked` - `integer`, лайкнул ли текущий пользователь, `0` или `1`, вычисляется - `liked`
`is_owner` - `integer`, текущий пользователь владелец записи, `0` или `1`, вычисляется - `is_owner`

<b>`likes`</b> - лайки
`id` - `bigint`, идентификатор лайка, первичный ключ - не отдаётся в JSON
`user_id` - `bigint`, кто поставил лайк, внешний ключ на `users` - не отдаётся в JSON
`planet_id` - `bigint`, чья планета, внешний ключ на `planets` - не отдаётся в JSON

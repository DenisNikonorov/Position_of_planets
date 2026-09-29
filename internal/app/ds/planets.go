package ds

import "time"

// Статусы планеты. Хранятся только в БД и никогда не отдаются клиенте:
// в интерфейсе это разные окна (черновик / публикация / удалено).
const (
	PlanetStatusDraft     = "draft"
	PlanetStatusPublished = "published"
	PlanetStatusDeleted   = "deleted"
)

// Planet — модель планеты, отображается в таблицу planets.
//
// Системные поля (ID, UserID, Status, CreatedAt, UpdatedAt) вычисляются
// на бэкенде и никогда не приходят от клиента.
type Planet struct {
	ID              uint    `gorm:"primaryKey"`
	UserID          uint    `gorm:"not null"`
	ImageURL        string  `gorm:"type:varchar(150)"`
	URLVideo        string  `gorm:"type:varchar(150)"`
	Name            string  `gorm:"type:varchar(50);not null"`
	Description     string  `gorm:"type:varchar(255)"`
	DistanceFromSun float64 `gorm:"not null;default:0"`
	Perihelion      float64 `gorm:"not null;default:0"`
	ObservationDate string  `gorm:"type:varchar(10)"`
	Status          string  `gorm:"type:varchar(20);not null;default:'draft'"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type User struct {
	ID          uint   `gorm:"primaryKey"`
	Login       string `gorm:"type:varchar(25);unique;not null"`
	Password    string `gorm:"type:varchar(100);not null"`
	IsModerator bool   `gorm:"type:boolean;default:false"`
}

type Like struct {
	ID       uint `gorm:"primaryKey"`
	UserID   uint `gorm:"not null;uniqueIndex:idx_user_planet"`
	PlanetID uint `gorm:"not null;uniqueIndex:idx_user_planet"`
}

// PlanetSerializer — единый набор полей для любого ответа с планетой.
// Набор полей не зависит от наличия данных: пустое значение остаётся
// пустой строкой или нулём, но не исчезает и не становится null.
type PlanetSerializer struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	ImageURL        string `json:"image_url"`
	URLVideo        string `json:"url_video"`
	DistanceFromSun string `json:"distance_from_sun"`
	Perihelion      string `json:"perihelion"`
	ObservationDate string `json:"observation_date"`
	LikesCount      int    `json:"likes_count"`
	Liked           int    `json:"liked"`
	IsOwner         int    `json:"is_owner"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type UserSerializer struct {
	ID    uint   `json:"id"`
	Login string `json:"login"`
}

// CreatePlanetRequest — тело POST /api/planets.
// Все поля приходят multipart-формой: name, description, числа, дата,
// плюс файлы под ключами image_url и url_video.
// Имя обязательно, остальные поля опциональны: не прислал — в JSON
// останется пустая строка или ноль.
type CreatePlanetRequest struct {
	Name            string  `form:"name" binding:"required"`
	Description     string  `form:"description"`
	DistanceFromSun float64 `form:"distance_from_sun"`
	Perihelion      float64 `form:"perihelion"`
	ObservationDate string  `form:"observation_date"`
}

// LikeRequest — тело POST /api/planets/:id/like. 1 ставит лайк, 0 отменяет.
// Указатель обязателен: тег required отбрасывает нулевое значение,
// а 0 — это валидное «отменить лайк».
type LikeRequest struct {
	Like *int `json:"like" binding:"required,oneof=0 1"`
}

// RegisterUserRequest — тело POST /api/users.
// id и is_moderator — системные поля, клиент их не передаёт.
type RegisterUserRequest struct {
	Login    string `json:"login" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=4"`
}

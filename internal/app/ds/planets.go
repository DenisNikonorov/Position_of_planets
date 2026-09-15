package ds

import "time"

type Planet struct {
	ID              int     `gorm:"primaryKey"`
	UserID          uint    `gorm:"not null;default:1"`
	ImageURL        string  `gorm:"type:varchar(150)"`
	URLVideo        string  `gorm:"type:varchar(150)"`
	Name            string  `gorm:"type:varchar(50);not null"`
	Description     string  `gorm:"type:varchar(255)"`
	DistanceFromSun float64 `gorm:"not null"`
	Perihelion      float64
	LikesCount      int    `gorm:"-"`
	HasVideo        bool   `gorm:"-"`
	Status          string `gorm:"type:varchar(20);not null;default:'draft'"`
	ObservationDate string `gorm:"type:varchar(10)"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Like struct {
	ID       uint `gorm:"primaryKey"`
	UserID   uint `gorm:"not null;uniqueIndex:idx_user_planet"`
	PlanetID uint `gorm:"not null;uniqueIndex:idx_user_planet"`

	User   User   `gorm:"foreignKey:UserID"`
	Planet Planet `gorm:"foreignKey:PlanetID"`
}

type User struct {
	ID          uint   `gorm:"primary_key" json:"id"`
	Login       string `gorm:"type:varchar(25);unique;not null" json:"login"`
	Password    string `gorm:"type:varchar(100);not null" json:"-"`
	IsModerator bool   `gorm:"type:boolean;default:false" json:"is_moderator"`
}

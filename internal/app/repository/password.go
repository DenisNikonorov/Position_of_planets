package repository

import (
	"crypto/subtle"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword хеширует пароль перед сохранением в БД.
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("хеширование пароля: %w", err)
	}
	return string(hashed), nil
}

// CheckPassword сверяет присланный пароль с тем, что лежит в БД.
// Записи, созданные до появления хеширования (пользователи из seed),
// хранят пароль в открытом виде — для них используется сравнение
// за постоянное время.
func CheckPassword(stored, password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)); err == nil {
		return true
	}
	return subtle.ConstantTimeCompare([]byte(stored), []byte(password)) == 1
}

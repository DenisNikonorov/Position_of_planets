package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"planets/internal/app/ds"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const bucketName = "planets"

func newMinIO() *minio.Client {
	minioClient, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("admin", "adminpassword", ""),
		Secure: false,
	})
	if err != nil {
		logrus.Warn("MinIO unavailable, media will use placeholders: ", err)
		return nil
	}
	return minioClient
}

// getMediaURL возвращает presigned URL для объекта MinIO,
// либо пустую строку, если медиа недоступно.
// UploadImage сохраняет загруженный файл в bucket MinIO и возвращает
// имя объекта и признак видео (по расширению .mp4).
func (r *Repository) UploadImage(file *multipart.FileHeader) (objectName string, hasVideo bool, err error) {
	if r.minio == nil {
		return "", false, nil
	}
	if file == nil {
		return "", false, nil
	}

	src, err := file.Open()
	if err != nil {
		return "", false, err
	}
	defer src.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := r.minio.BucketExists(ctx, bucketName)
	if err != nil {
		return "", false, err
	}
	if !exists {
		if err := r.minio.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			return "", false, err
		}
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	objectName = uuid.NewString() + ext

	_, err = r.minio.PutObject(ctx, bucketName, objectName, src, file.Size, minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")})
	if err != nil {
		return "", false, err
	}

	hasVideo = ext == ".mp4"
	return objectName, hasVideo, nil
}

func (r *Repository) getMediaURL(objectName string) string {
	if r.minio == nil {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	exists, err := r.minio.BucketExists(ctx, bucketName)
	if err != nil || !exists {
		return ""
	}

	reqParams := make(url.Values)
	presignedURL, err := r.minio.PresignedGetObject(ctx, bucketName, objectName, 24*time.Hour, reqParams)
	if err != nil {
		logrus.Warn("Failed to presign object ", objectName, ": ", err)
		return ""
	}
	return presignedURL.String()
}

const (
	PlanetStatusPublished = "published"
	PlanetStatusDraft     = "draft"
	PlanetStatusDeleted   = "deleted"
)

// resolveMedia подменяет имена объектов MinIO на presigned URL.
// Если MinIO недоступен - очищаем URL, чтобы шаблоны показали плейсхолдер.
// Признак видео определяется наличием ссылки в столбце url_video.
func (r *Repository) resolveMedia(planet *ds.Planet) {
	planet.HasVideo = planet.URLVideo != ""
	if planet.URLVideo != "" {
		planet.URLVideo = r.getMediaURL(planet.URLVideo)
	}
	if planet.ImageURL != "" {
		planet.ImageURL = r.getMediaURL(planet.ImageURL)
	}
}

type likesCountRow struct {
	PlanetID int
	Count    int64
}

// attachLikesCount заполняет LikesCount планет реальным количеством записей
// из таблицы likes (подсчётом) — колонки likes_count в базе нет.
func (r *Repository) attachLikesCount(planets []ds.Planet) error {
	if len(planets) == 0 {
		return nil
	}

	var rows []likesCountRow
	err := r.db.Model(&ds.Like{}).
		Select("planet_id, COUNT(*) AS count").
		Group("planet_id").
		Scan(&rows).Error
	if err != nil {
		return err
	}

	counts := make(map[int]int64, len(rows))
	for _, row := range rows {
		counts[row.PlanetID] = row.Count
	}

	for i := range planets {
		planets[i].LikesCount = int(counts[planets[i].ID])
	}

	return nil
}

func (r *Repository) findPlanets(query string, args ...interface{}) ([]ds.Planet, error) {
	var planets []ds.Planet
	err := r.db.Where(query, args...).Find(&planets).Error
	if err != nil {
		return nil, err
	}
	if err := r.attachLikesCount(planets); err != nil {
		return nil, err
	}
	for i := range planets {
		r.resolveMedia(&planets[i])
	}
	return planets, nil
}

func (r *Repository) GetAllPlanets() ([]ds.Planet, error) {
	return r.findPlanets("status = ?", PlanetStatusPublished)
}

func (r *Repository) GetPlanetByID(id int) (*ds.Planet, error) {
	// используем "курсор" - сырой запрос со строковым указателем на результат
	query := "SELECT id, image_url, url_video, name, description, distance_from_sun, perihelion, observation_date FROM planets WHERE id = $1 AND status = '" + PlanetStatusPublished + "'"

	row := r.db.Raw(query, id).Row()

	planet := &ds.Planet{}
	err := row.Scan(
		&planet.ID,
		&planet.ImageURL,
		&planet.URLVideo,
		&planet.Name,
		&planet.Description,
		&planet.DistanceFromSun,
		&planet.Perihelion,
		&planet.ObservationDate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var likesCount int64
	if err := r.db.Model(&ds.Like{}).Where("planet_id = ?", id).Count(&likesCount).Error; err != nil {
		return nil, err
	}
	planet.LikesCount = int(likesCount)

	r.resolveMedia(planet)

	return planet, nil
}

func (r *Repository) SearchPlanetsByName(name string) ([]ds.Planet, error) {
	return r.findPlanets("name ILIKE ? AND status = ?", "%"+name+"%", PlanetStatusPublished)
}

// FindPlanetByName ищет существующую карточку по точному названию без учёта регистра.
// Если карточка найдена — добавление идёт сразу на шаг 2 с подставленными данными.
func (r *Repository) FindPlanetByName(name string) (*ds.Planet, error) {
	var planet ds.Planet
	err := r.db.Where("LOWER(name) = LOWER(?) AND status <> ?", name, PlanetStatusDeleted).First(&planet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &planet, nil
}

func (r *Repository) GetPlanetsByMaxDistance(maxDist float64) ([]ds.Planet, error) {
	return r.findPlanets("distance_from_sun <= ? AND status = ?", maxDist, PlanetStatusPublished)
}

// CreatePlanetDraft создаёт черновик планеты (шаг 1) и возвращает его id.
func (r *Repository) CreatePlanetDraft(name, imageURL, urlVideo string) (int, error) {
	planet := ds.Planet{
		UserID:   1,
		ImageURL: imageURL,
		URLVideo: urlVideo,
		Name:     name,
		Status:   PlanetStatusDraft,
	}
	if err := r.db.Create(&planet).Error; err != nil {
		return 0, fmt.Errorf("ошибка при создании черновика планеты: %w", err)
	}
	return planet.ID, nil
}

// PublishPlanet заполняет данные шага 2 и переводит черновик в опубликованный.
func (r *Repository) PublishPlanet(id int, description string, distance float64, observationDate string) error {
	err := r.db.Model(&ds.Planet{}).
		Where("id = ? AND status <> ?", id, PlanetStatusDeleted).
		Updates(map[string]interface{}{
			"description":       description,
			"distance_from_sun": distance,
			"observation_date":  observationDate,
			"status":            PlanetStatusPublished,
		}).Error
	if err != nil {
		return fmt.Errorf("ошибка при публикации планеты с id %d: %w", id, err)
	}
	return nil
}

// GetPlanetForEdit возвращает планету без фильтра статуса (для шага 2 черновика).
func (r *Repository) GetPlanetForEdit(id int) (*ds.Planet, error) {
	var planet ds.Planet
	err := r.db.Where("id = ? AND status <> ?", id, PlanetStatusDeleted).First(&planet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	r.resolveMedia(&planet)
	return &planet, nil
}

// GetFirstDraft возвращает самый свежий черновик, чтобы шаг 1 переходил сразу на шаг 2.
func (r *Repository) GetFirstDraft() (*ds.Planet, error) {
	var planet ds.Planet
	err := r.db.Where("status = ?", PlanetStatusDraft).
		Order("id DESC").
		First(&planet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &planet, nil
}

func (r *Repository) DeletePlanet(planetID uint) error {
	err := r.db.Model(&ds.Planet{}).Where("id = ?", planetID).UpdateColumn("status", PlanetStatusDeleted).Error
	if err != nil {
		return fmt.Errorf("ошибка при удалении планеты с id %d: %w", planetID, err)
	}

	return nil
}

// GetUserLikesCount возвращает количество планет, отмеченных лайком пользователя.
// В дальнейшем вместо хардкода id пользователя будет авторизация.
func (r *Repository) GetUserLikesCount(userID uint) int64 {
	var count int64
	err := r.db.Model(&ds.Like{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting likes:", err)
	}

	return count
}

type Repository struct {
	db    *gorm.DB
	minio *minio.Client
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	minioClient := newMinIO()

	return &Repository{
		db:    db,
		minio: minioClient,
	}, nil
}

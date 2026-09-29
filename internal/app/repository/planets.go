package repository

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"planets/internal/app/ds"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Ошибки репозитория. Обработчики превращают их в коды состояния,
// текст ошибки клиенту не отдаётся.
var (
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("forbidden")
)

// Дефолтные медиафайлы: лежат в репозитории и заливаются в MinIO при старте,
// поэтому ссылка на них всегда рабочая.
const (
	DefaultPhotoName = "default_photo.png"
	DefaultVideoName = "default_video.mp4"
)

type RepositorySettings struct {
	PostgresDSN    string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool
	MinioPublicURL string
}

type Repository struct {
	db             *gorm.DB
	minio          *minio.Client
	bucket         string
	minioPublicURL string
}

// New создаёт подключения к БД и MinIO.
func New(settings RepositorySettings) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(settings.PostgresDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к БД: %w", err)
	}

	minioClient, err := minio.New(settings.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(settings.MinioAccessKey, settings.MinioSecretKey, ""),
		Secure: settings.MinioUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к MinIO: %w", err)
	}

	return &Repository{
		db:             db,
		minio:          minioClient,
		bucket:         settings.MinioBucket,
		minioPublicURL: strings.TrimSuffix(settings.MinioPublicURL, "/"),
	}, nil
}

// mediaURL собирает публичную ссылку на объект в бакете. Бакет открыт
// на чтение, поэтому ссылки стабильные и не протухают, как presigned.
func (r *Repository) mediaURL(objectName string) string {
	if objectName == "" {
		return ""
	}
	return r.minioPublicURL + "/" + r.bucket + "/" + objectName
}

// EnsureDefaults создаёт бакет и кладёт в него дефолтные фото и видео,
// если их там ещё нет. Вызывается один раз при старте приложения.
func (r *Repository) EnsureDefaults(defaults map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exists, err := r.minio.BucketExists(ctx, r.bucket)
	if err != nil {
		return fmt.Errorf("проверка бакета: %w", err)
	}
	if !exists {
		if err := r.minio.MakeBucket(ctx, r.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("создание бакета: %w", err)
		}
	}

	for objectName, path := range defaults {
		if _, err := r.minio.StatObject(ctx, r.bucket, objectName, minio.StatObjectOptions{}); err == nil {
			continue
		}
		if err := r.uploadFile(ctx, path, objectName); err != nil {
			return err
		}
	}
	return nil
}

// uploadFile кладёт локальный файл в бакет под заданным именем.
func (r *Repository) uploadFile(ctx context.Context, path, objectName string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("открытие файла %s: %w", path, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("чтение файла %s: %w", path, err)
	}

	if _, err := r.minio.PutObject(ctx, r.bucket, objectName, file, info.Size(), minio.PutObjectOptions{
		ContentType: contentTypeByExt(filepath.Ext(path)),
	}); err != nil {
		return fmt.Errorf("загрузка в MinIO %s: %w", objectName, err)
	}
	return nil
}

func contentTypeByExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".mp4":
		return "video/mp4"
	default:
		return "application/octet-stream"
	}
}

// UploadMedia сохраняет загруженный файл в MinIO и возвращает его имя.
// Имя генерируется на латинице и кладётся в поле image_url / url_video.
func (r *Repository) UploadMedia(header *multipart.FileHeader) (string, error) {
	if header == nil {
		return "", nil
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		return "", ErrForbidden
	}

	src, err := header.Open()
	if err != nil {
		return "", fmt.Errorf("открытие загруженного файла: %w", err)
	}
	defer src.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectName := uuid.NewString() + ext
	if _, err := r.minio.PutObject(ctx, r.bucket, objectName, src, header.Size, minio.PutObjectOptions{
		ContentType: contentTypeByExt(ext),
	}); err != nil {
		return "", fmt.Errorf("загрузка в MinIO: %w", err)
	}

	return objectName, nil
}

// likesStats возвращает счётчики лайков и множество планет,
// которые текущий пользователь уже лайкнул.
func (r *Repository) likesStats(planetIDs []uint, userID uint) (map[uint]int, map[uint]bool, error) {
	counts := make(map[uint]int, len(planetIDs))
	liked := make(map[uint]bool, len(planetIDs))
	if len(planetIDs) == 0 {
		return counts, liked, nil
	}

	var rows []ds.Like
	if err := r.db.Where("planet_id IN ?", planetIDs).Find(&rows).Error; err != nil {
		return nil, nil, fmt.Errorf("чтение лайков: %w", err)
	}

	for _, row := range rows {
		counts[row.PlanetID]++
		if row.UserID == userID {
			liked[row.PlanetID] = true
		}
	}
	return counts, liked, nil
}

func (r *Repository) serialize(planets []ds.Planet, userID uint) []ds.PlanetSerializer {
	ids := make([]uint, 0, len(planets))
	for _, p := range planets {
		ids = append(ids, p.ID)
	}

	counts, liked, err := r.likesStats(ids, userID)
	if err != nil {
		counts = map[uint]int{}
		liked = map[uint]bool{}
	}

	result := make([]ds.PlanetSerializer, 0, len(planets))
	for _, p := range planets {
		result = append(result, ds.PlanetSerializer{
			ID:              p.ID,
			Name:            p.Name,
			Description:     p.Description,
			ImageURL:        r.mediaURL(p.ImageURL),
			URLVideo:        r.mediaURL(p.URLVideo),
			DistanceFromSun: formatFloat(p.DistanceFromSun),
			Perihelion:      formatFloat(p.Perihelion),
			ObservationDate: p.ObservationDate,
			LikesCount:      counts[p.ID],
			Liked:           boolToInt(liked[p.ID]),
			IsOwner:         boolToInt(p.UserID == userID),
			CreatedAt:       formatTime(p.CreatedAt),
			UpdatedAt:       formatTime(p.UpdatedAt),
		})
	}
	return result
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

// GetPlanets — список опубликованных планет с фильтрацией на бэкенде.
// Фильтры: search по имени (ILIKE), maxDist по расстоянию от Солнца.
// Удалённые записи клиенту не передаются.
func (r *Repository) GetPlanets(search string, maxDist *float64, userID uint) ([]ds.PlanetSerializer, error) {
	var planets []ds.Planet

	query := r.db.Model(&ds.Planet{}).Where("status = ?", ds.PlanetStatusPublished)
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	if maxDist != nil {
		query = query.Where("distance_from_sun <= ?", *maxDist)
	}

	if err := query.Order("id").Find(&planets).Error; err != nil {
		return nil, fmt.Errorf("чтение списка планет: %w", err)
	}
	return r.serialize(planets, userID), nil
}

// GetFeed — лента опубликованных планет, листание по кругу.
// Без id возвращает первую. С id и next=true — следующую после id
// с переходом на первую, если id был последним.
func (r *Repository) GetFeed(id *uint, next bool, userID uint) (*ds.PlanetSerializer, error) {
	var planets []ds.Planet
	if err := r.db.Where("status = ?", ds.PlanetStatusPublished).Order("id").Find(&planets).Error; err != nil {
		return nil, fmt.Errorf("чтение ленты: %w", err)
	}
	if len(planets) == 0 {
		return nil, ErrNotFound
	}

	index := 0
	if id != nil {
		found := -1
		for i, p := range planets {
			if p.ID == *id {
				found = i
				break
			}
		}
		if found == -1 {
			return nil, ErrNotFound
		}
		if next {
			index = (found + 1) % len(planets)
		} else {
			index = found
		}
	}

	serialized := r.serialize(planets[index:index+1], userID)
	return &serialized[0], nil
}

// GetUserDraft — черновик текущего пользователя, не более одной записи.
// Идентификатор не передаётся: он определяется пользователем.
func (r *Repository) GetUserDraft(userID uint) (*ds.PlanetSerializer, error) {
	var planet ds.Planet
	err := r.db.Where("user_id = ? AND status = ?", userID, ds.PlanetStatusDraft).
		Order("id DESC").First(&planet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("чтение черновика: %w", err)
	}

	serialized := r.serialize([]ds.Planet{planet}, userID)
	return &serialized[0], nil
}

// CreatePlanet создаёт запись в статусе draft от имени userID.
// Статус и создатель вычисляются на бэкенде; бизнес-поля берутся из формы.
func (r *Repository) CreatePlanet(userID uint, req ds.CreatePlanetRequest, imageName, videoName string) (*ds.PlanetSerializer, error) {
	planet := ds.Planet{
		UserID:          userID,
		Name:            req.Name,
		Description:     req.Description,
		DistanceFromSun: req.DistanceFromSun,
		Perihelion:      req.Perihelion,
		ObservationDate: req.ObservationDate,
		ImageURL:        imageName,
		URLVideo:        videoName,
		Status:          ds.PlanetStatusDraft,
	}

	if err := r.db.Create(&planet).Error; err != nil {
		return nil, fmt.Errorf("создание черновика: %w", err)
	}

	serialized := r.serialize([]ds.Planet{planet}, userID)
	return &serialized[0], nil
}

// PublishPlanet переводит черновик в опубликованный.
// Переход разрешён только из draft: обратно в черновик вернуться нельзя.
// Поля планеты при этом не изменяются.
func (r *Repository) PublishPlanet(userID uint, id uint) (*ds.PlanetSerializer, error) {
	result := r.db.Model(&ds.Planet{}).
		Where("id = ? AND user_id = ? AND status = ?", id, userID, ds.PlanetStatusDraft).
		Update("status", ds.PlanetStatusPublished)
	if result.Error != nil {
		return nil, fmt.Errorf("публикация планеты: %w", result.Error)
	}
	// Ничего не обновилось: записи нет, она чужая или уже опубликована.
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}

	planet, err := r.findPlanet(id, userID, ds.PlanetStatusPublished)
	if err != nil {
		return nil, err
	}

	serialized := r.serialize([]ds.Planet{*planet}, userID)
	return &serialized[0], nil
}

// SoftDeletePlanet помечает планету удалённой. Физически запись остаётся
// в БД и клиенту не передаётся. Удалить можно только свою запись,
// и только один раз: повторный вызов не меняет ни одной строки.
func (r *Repository) SoftDeletePlanet(userID uint, id uint) (*ds.PlanetSerializer, error) {
	result := r.db.Model(&ds.Planet{}).
		Where("id = ? AND user_id = ? AND status <> ?", id, userID, ds.PlanetStatusDeleted).
		Update("status", ds.PlanetStatusDeleted)
	if result.Error != nil {
		return nil, fmt.Errorf("удаление планеты: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}

	planet, err := r.findPlanetIncludingDeleted(id, userID)
	if err != nil {
		return nil, err
	}

	serialized := r.serialize([]ds.Planet{*planet}, userID)
	return &serialized[0], nil
}

func (r *Repository) findPlanet(id uint, userID uint, status string) (*ds.Planet, error) {
	var planet ds.Planet
	err := r.db.Where("id = ? AND user_id = ? AND status = ?", id, userID, status).First(&planet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("чтение планеты: %w", err)
	}
	return &planet, nil
}

func (r *Repository) findPlanetIncludingDeleted(id uint, userID uint) (*ds.Planet, error) {
	var planet ds.Planet
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&planet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("чтение планеты: %w", err)
	}
	return &planet, nil
}

// findPublishedPlanet читает опубликованную планету независимо от владельца:
// лайк может поставить любой пользователь.
func (r *Repository) findPublishedPlanet(id uint) (*ds.Planet, error) {
	var planet ds.Planet
	err := r.db.Where("id = ? AND status = ?", id, ds.PlanetStatusPublished).First(&planet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("чтение планеты: %w", err)
	}
	return &planet, nil
}

// SetLike ставит или снимает лайк текущего пользователя.
// like = 1 ставит лайк, like = 0 отменяет его.
func (r *Repository) SetLike(userID uint, planetID uint, like int) (*ds.PlanetSerializer, error) {
	var count int64
	if err := r.db.Model(&ds.Planet{}).
		Where("id = ? AND status = ?", planetID, ds.PlanetStatusPublished).
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("проверка планеты: %w", err)
	}
	if count == 0 {
		return nil, ErrNotFound
	}

	if like == 1 {
		var existing int64
		if err := r.db.Model(&ds.Like{}).
			Where("user_id = ? AND planet_id = ?", userID, planetID).
			Count(&existing).Error; err != nil {
			return nil, fmt.Errorf("проверка лайка: %w", err)
		}
		if existing == 0 {
			if err := r.db.Create(&ds.Like{UserID: userID, PlanetID: planetID}).Error; err != nil {
				return nil, fmt.Errorf("установка лайка: %w", err)
			}
		}
	} else {
		if err := r.db.Where("user_id = ? AND planet_id = ?", userID, planetID).
			Delete(&ds.Like{}).Error; err != nil {
			return nil, fmt.Errorf("отмена лайка: %w", err)
		}
	}

	planet, err := r.findPublishedPlanet(planetID)
	if err != nil {
		return nil, err
	}

	serialized := r.serialize([]ds.Planet{*planet}, userID)
	return &serialized[0], nil
}

// CreateUser регистрирует нового пользователя. id и is_moderator —
// системные поля, клиент их не передаёт.
func (r *Repository) CreateUser(login, password string) (*ds.UserSerializer, error) {
	var existing int64
	if err := r.db.Model(&ds.User{}).Where("login = ?", login).Count(&existing).Error; err != nil {
		return nil, fmt.Errorf("проверка логина: %w", err)
	}
	if existing > 0 {
		return nil, ErrForbidden
	}

	user := ds.User{Login: login, Password: password, IsModerator: false}
	if err := r.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("создание пользователя: %w", err)
	}

	return &ds.UserSerializer{ID: user.ID, Login: user.Login}, nil
}

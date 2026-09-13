package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

type Database struct {
	MinIO *minio.Client
}

func NewDB() (*Database, error) {
	minioClient, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("admin", "adminpassword", ""),
		Secure: false,
	})
	if err != nil {
		logrus.Warn("MinIO unavailable, media will use placeholders: ", err)
		return &Database{MinIO: nil}, nil
	}
	return &Database{MinIO: minioClient}, nil
}

type Planet struct {
	ID              int
	Name            string
	Description     string
	DistanceFromSun float64
	LikesCount      int
	Perihelion      float64
	ImageURL        string
	HasVideo        bool
}

const bucketName = "planets"

func (d *Database) getMediaURL(objectName string) string {
	if d.MinIO == nil {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	exists, err := d.MinIO.BucketExists(ctx, bucketName)
	if err != nil || !exists {
		return ""
	}

	reqParams := make(url.Values)
	presignedURL, err := d.MinIO.PresignedGetObject(ctx, bucketName, objectName, 24*time.Hour, reqParams)
	if err != nil {
		logrus.Warn("Failed to presign object ", objectName, ": ", err)
		return ""
	}
	return presignedURL.String()
}

func (d *Database) GetData() ([]Planet, error) {
	planets := []Planet{
		{
			ID:              1,
			Name:            "Меркурий",
			Description:     "Ближайшая к Солнцу планета. Поверхность покрыта кратерами, атмосфера практически отсутствует. Температура колеблется от -180°C ночью до +430°C днём.",
			DistanceFromSun: 0.39,
			LikesCount:      412,
			Perihelion:      46.0,
			ImageURL:        "",
			HasVideo:        false,
		},
		{
			ID:              2,
			Name:            "Земля",
			Description:     "Единственная известная планета с жидкой водой и жизнью. Атмосфера состоит преимущественно из азота и кислорода. Имеет один естественный спутник — Луну.",
			DistanceFromSun: 1.00,
			LikesCount:      1200,
			Perihelion:      102.9,
			ImageURL:        d.getMediaURL("Earth.png"),
			HasVideo:        false,
		},
		{
			ID:              3,
			Name:            "Марс",
			Description:     "Красная планета с разреженной атмосферой. Имеет два небольших спутника — Фобос и Деймос. Ржавый цвет поверхности обусловлен оксидом железа.",
			DistanceFromSun: 1.52,
			LikesCount:      956,
			Perihelion:      336.1,
			ImageURL:        d.getMediaURL("Mars_spin.mp4"),
			HasVideo:        true,
		},
		{
			ID:              4,
			Name:            "Юпитер",
			Description:     "Крупнейшая планета Солнечной системы. Газовый гигант с мощной системой вихрей, включая Большое Красное Пятно — шторм, длящийся сотни лет.",
			DistanceFromSun: 5.20,
			LikesCount:      2100,
			Perihelion:      14.7,
			ImageURL:        d.getMediaURL("Jupiter_spin.mp4"),
			HasVideo:        true,
		},
		{
			ID:              5,
			Name:            "Сатурн",
			Description:     "Планета с впечатляющей системой колец из льда и камня. Плотность Сатурна меньше воды — он бы всплыл в достаточно большом океане.",
			DistanceFromSun: 9.58,
			LikesCount:      1800,
			Perihelion:      95.5,
			ImageURL:        d.getMediaURL("Saturn.png"),
			HasVideo:        false,
		},
	}
	if len(planets) == 0 {
		return nil, fmt.Errorf("empty array")
	}
	return planets, nil
}

func (d *Database) GetPlanetByID(id int) (Planet, error) {
	planets, err := d.GetData()
	if err != nil {
		return Planet{}, err
	}
	for _, p := range planets {
		if p.ID == id {
			return p, nil
		}
	}
	return Planet{}, fmt.Errorf("planet with ID %d not found", id)
}

func (d *Database) GetPlanetsByMaxDistance(maxDist float64) ([]Planet, error) {
	planets, err := d.GetData()
	if err != nil {
		return nil, err
	}
	var result []Planet
	for _, p := range planets {
		if p.DistanceFromSun <= maxDist {
			result = append(result, p)
		}
	}
	return result, nil
}

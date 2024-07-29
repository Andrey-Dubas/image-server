package image_repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type ImageMetadata struct {
	Filename string
	Format   string
	Camera   string
}

type RepositoryStatistics struct {
	MostPopularFormat  string
	MostPopularCameras []string
	UploadFrequency    map[string]int
}

type IImageRepository interface {
	SaveImageMetadata(ImageMetadata) error
	GetImageMetadata(string) (*ImageMetadata, error)
	GetStatistics() (*RepositoryStatistics, error)
}

type PostgresImageRepository struct {
	postgresClient *sqlx.DB
}

func NewPostgresImageRepository(host string, password string, user string, dbname string) (IImageRepository, error) {
	db, err := sqlx.Connect("postgres", fmt.Sprintf("host=%s port=5432 dbname=%s user=%s password=%s sslmode=disable", host, dbname, user, password))
	if err != nil {
		return nil, err
	}
	return PostgresImageRepository{postgresClient: db}, nil
}

func (r PostgresImageRepository) SaveImageMetadata(data ImageMetadata) error {
	rows, err := r.postgresClient.Queryx(
		"INSERT INTO app.images(filename, camera_model, format) VALUES(?, ?, ?)",
		data.Filename,
		data.Camera,
		data.Format,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	return err
}

func (r PostgresImageRepository) GetImageMetadata(filename string) (*ImageMetadata, error) {
	rows, err := r.postgresClient.Queryx(
		"SELECT filename, camera_model, format FROM app.images WHERE filename = %1", filename,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var returnedFilename, camera, format string
	err = rows.Scan(&returnedFilename, &camera, &format)
	if err != nil {
		return nil, err
	}

	return &ImageMetadata{
		Filename: returnedFilename,
		Format:   format,
		Camera:   camera,
	}, nil
}

func (r PostgresImageRepository) GetStatistics() (*RepositoryStatistics, error) {
	var format string
	formatRows, err := r.postgresClient.Queryx(`
	SELECT format, COUNT(format)
	GROUP BY format
	ORDER BY COUNT(format) DESC
	LIMIT 1
	FROM app.images
	`)
	if err != nil {
		return nil, err
	}

	err = formatRows.Scan(&format)
	if err != nil {
		return nil, err
	}

	modelRows, err := r.postgresClient.Queryx(`
	SELECT model, COUNT(model)
	GROUP BY model
	ORDER BY COUNT(model) DESC
	LIMIT 10
	FROM app.images
	`)
	if err != nil {
		return nil, err
	}
	defer modelRows.Close()

	models := make([]string, 0, 10)
	var m string
	for modelRows.Next() {
		modelRows.Scan(&m)
		models = append(models, m)
	}

	frequencyUploadRows, err := r.postgresClient.Queryx(`
	SELECT date(created_at), COUNT(date(created_at)) as frequency
	WHERE created_at >= NOW() - INTERVAL '30 DAY'
	FROM app.images
	`)
	if err != nil {
		return nil, err
	}
	defer frequencyUploadRows.Close()

	date := ""
	frequency := 0
	frequencyMap := make(map[string]int)

	for frequencyUploadRows.Next() {
		modelRows.Scan(&date, &frequency)
		models = append(models, m)
	}

	return &RepositoryStatistics{
		MostPopularFormat:  format,
		MostPopularCameras: models,
		UploadFrequency:    frequencyMap,
	}, nil
}

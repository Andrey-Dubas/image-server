package persistent_storage

import (
	"io"
	"mime/multipart"
	"os"

)

type IStorage interface {
	UploadImage(file multipart.File, filename string) error
	GetImage(filename string) (io.Reader, error)
}

type LocalImageRepository struct {
	rootPath string
}

func newLocalImageRepository(rootPath string) IStorage {
	return LocalImageRepository{
		rootPath: rootPath,
	}
}

func (s LocalImageRepository) UploadImage(file multipart.File, filename string) error {
	dst, err := os.Create(s.rootPath + "/" + filename)
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, file)
	if err != nil {
		return err
	}

	return nil
}

func (s LocalImageRepository) GetImage(filename string) (io.Reader, error) {
	return os.Open(s.rootPath + "/" + filename)
}

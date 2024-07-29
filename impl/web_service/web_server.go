package web_service

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rwcarlsen/goexif/exif"

	image_repository "github.com/Andrey-Dubas/image-server/image_repository"
	link_repository "github.com/Andrey-Dubas/image-server/link_repository"
	storage "github.com/Andrey-Dubas/image-server/storage"
)

type GenerateUploadLinkRequest struct {
	TTL time.Duration `json:"ttl"`
}

type GenerateUploadLinkResponse struct {
	Link string `json:"link"`
}

type UploadLinkResponse GenerateUploadLinkResponse

type WebService struct {
	LinkRepo  link_repository.ILinkRepository
	ImageRepo image_repository.IImageRepository
	Storage   storage.IStorage
	Token     string
	SelfUrl   string
}

func newWebService(
	linkRepository link_repository.ILinkRepository,
	imageRepository image_repository.IImageRepository,
	storage storage.IStorage,
	selfUrl string) WebService {

	return WebService{
		LinkRepo:  linkRepository,
		ImageRepo: imageRepository,
		Storage:   storage,
		SelfUrl:   selfUrl,
	}
}

func (s WebService) VerifyTokenFunction() func(c *gin.Context) {
	return func(c *gin.Context) {
		if c.GetHeader("token") == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provide token"})
		}
		if s.Token != c.GetHeader("token") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		} else {
			c.Next()
		}
	}
}

func (s WebService) generateLink() func(c *gin.Context) {
	return func(c *gin.Context) {
		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Error(err)
			return
		}

		var request GenerateUploadLinkRequest
		err = json.Unmarshal(data, &request)
		if err != nil {
			c.Error(err)
			return
		}
		uuid, err := s.LinkRepo.GenerateLink(request.TTL)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusCreated, GenerateUploadLinkResponse{
			Link: s.SelfUrl + "/upload/image/" + uuid.String() + "/",
		})

	}
}

func (s WebService) handleImageUpload() func(c *gin.Context) {
	return func(c *gin.Context) {
		file, header, err := c.Request.FormFile("image")

		if err != nil {
			c.Error(err)
			return
		}

		fileExt := filepath.Ext(header.Filename)
		currentPath := strings.Split(c.Request.URL.Path, "/")
		uuid := currentPath[len(currentPath)-1]
		originalFileName := strings.TrimSuffix(
			filepath.Base(header.Filename),
			filepath.Ext(header.Filename)) +
			"." + uuid +
			"." + fileExt

		filename := strings.ReplaceAll(strings.ToLower(originalFileName), " ", "-")

		err = s.Storage.UploadImage(file, filename)
		if err != nil {
			c.Error(err)
		}

		imageMetadata, err := exif.Decode(file)
		if err != nil {
			c.Error(err)
		}

		model, err := imageMetadata.Get(exif.Model)
		if err != nil {
			c.Error(err)
		}

		err = s.ImageRepo.SaveImageMetadata(
			image_repository.ImageMetadata{
				Filename: filename,
				Format:   fileExt,
				Camera:   model.String(),
			},
		)
		if err != nil {
			c.Error(err)
		}

		c.JSON(http.StatusCreated, UploadLinkResponse{
			Link: s.SelfUrl + "/images/" + filename,
		})
	}
}

func (s WebService) getImage() func(c *gin.Context) {
	return func(c *gin.Context) {
		splitUrl := strings.Split(c.Request.URL.Path, "/")
		fileID := splitUrl[len(splitUrl)-1]

		reader, err := s.Storage.GetImage(fileID)
		if err != nil {
			c.Error(err)
		}

		c.Header("Content-Disposition", "attachment; filename="+fileID)
		c.Header("Content-Type", "application/octet-stream")
		c.Stream(func(w io.Writer) bool {
			_, err := io.Copy(w, reader)
			if err != nil {
				c.Error(err)
				return true
			}
			return false
		})
	}
}

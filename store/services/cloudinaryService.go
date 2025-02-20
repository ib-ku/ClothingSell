package services

import (
	"context"
	"log"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryService(cloudName, apiKey, apiSecret string) *CloudinaryService {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		log.Fatal("Ошибка Cloudinary:", err)
	}
	return &CloudinaryService{cld: cld}
}

func (c *CloudinaryService) UploadImage(filePath string) (string, error) {
	uploadResult, err := c.cld.Upload.Upload(context.Background(), filePath, uploader.UploadParams{})
	if err != nil {
		return "", err
	}
	return uploadResult.SecureURL, nil
}

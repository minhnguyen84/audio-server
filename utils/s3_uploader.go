package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
)

type FileUploader interface {
	UploadFile(ctx context.Context, file *os.File, s string) error
}

// S3Uploader gère les uploads vers S3 like
type S3Uploader struct {
	s3Client   *s3.Client
	bucketName string
}

func NewS3Uploader(appConfig AppConfig) (*S3Uploader, error) {
	var s3Client *s3.Client

	if appConfig.S3Endpoint != "" {
		cfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(appConfig.S3Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(appConfig.S3AccessKey, appConfig.S3SecretKey, "")),
			config.WithBaseEndpoint(appConfig.S3Endpoint),
			config.WithHTTPClient(&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: !appConfig.S3UseSSL, // Désactiver la vérification SSL si non utilisé
					},
				},
			}),
		)

		if err != nil {
			log.Fatalf("Impossible de charger la configuration: %v", err)
		}
		// Créer un client S3 avec UsePathStyle
		s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.UsePathStyle = true // Nécessaire pour MinIO ou autre S3_like
		})
	} else {
		// Charger la configuration AWS
		cfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(appConfig.S3Region),
		)
		if err != nil {
			log.Fatalf("Impossible de charger la configuration: %v", err)
		}
		// Configuration pour AWS S3 avec rôles IAM
		s3Client = s3.NewFromConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("Erreur lors de la création de la session AWS: %v", err)
		}
	}

	if appConfig.S3BucketName == "" {
		return nil, fmt.Errorf("S3BucketName ne peut pas être null ou vide")
	}

	return &S3Uploader{
		s3Client:   s3Client,
		bucketName: appConfig.S3BucketName,
	}, nil
}

func (uploader *S3Uploader) UploadFile(ctx context.Context, file *os.File, s3Key string) error {
	// Réinitialiser le pointeur du fichier au début
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("erreur lors de la réinitialisation du pointeur du fichier: %w", err)
	}
	_, err := uploader.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(uploader.bucketName),
		Key:    aws.String(s3Key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("Erreur lors de l'upload vers S3 - %s: %v", uploader.bucketName, err)
	}

	return nil
}

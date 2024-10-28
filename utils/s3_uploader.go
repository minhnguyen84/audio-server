package utils

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3Uploader gère les uploads vers S3 like
type S3Uploader struct {
	s3Client   *s3.S3
	bucketName string
}

func NewS3Uploader(config AppConfig) (*S3Uploader, error) {
	var sess *session.Session
	var err error

	if config.S3Endpoint != "" {
		creds := credentials.NewStaticCredentials(config.S3AccessKey, config.S3SecretKey, "")
		_, err = creds.Get()
		if err != nil {
			return nil, fmt.Errorf("Erreur de récupération des credentials S3: %v", err)
		}

		sess, err = session.NewSession(&aws.Config{
			Region:           aws.String(config.S3Region),
			Endpoint:         aws.String(config.S3Endpoint),
			Credentials:      creds,
			S3ForcePathStyle: aws.Bool(true), // Nécessaire pour MinIO ou autre S3_like
			DisableSSL:       aws.Bool(!config.S3UseSSL),
		})
		if err != nil {
			return nil, fmt.Errorf("Erreur lors de la création de la session S3: %v", err)
		}
	} else {
		// Configuration pour AWS S3 avec rôles IAM
		sess, err = session.NewSession(&aws.Config{
			Region: aws.String(config.S3Region),
		})
		if err != nil {
			return nil, fmt.Errorf("Erreur lors de la création de la session AWS: %v", err)
		}
	}

	s3Client := s3.New(sess)

	return &S3Uploader{
		s3Client:   s3Client,
		bucketName: config.S3BucketName,
	}, nil
}

func (uploader *S3Uploader) UploadFile(ctx context.Context, file *os.File, s3Key string) error {
	_, err := uploader.s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(uploader.bucketName),
		Key:    aws.String(s3Key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("Erreur lors de l'upload vers S3: %v", err)
	}

	return nil
}

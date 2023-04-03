package imagefilteraccess

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/okpalaChidiebere/chirper-app-api-image/v0/common"
)


type S3PresignerRepository struct {
	presignClient common.PresignClientAPI
}

func NewPresignerRepository(client common.PresignClientAPI) *S3PresignerRepository {
	return &S3PresignerRepository{
		presignClient: client,
	}
}

func (r *S3PresignerRepository) GetGetSignedUrl(ctx context.Context, bucketName, objectKey string) (string, error) {
	request, err := r.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(120 * int64(time.Second))
	})
	if err != nil {
		log.Printf("Couldn't get a presigned request to get %v:%v. Here's why: %v\n",
			bucketName, objectKey, err)
	}
	return request.URL, err
}

func (r *S3PresignerRepository) GetPutSignedUrl(ctx context.Context, bucketName, objectKey string) (string, error) {
	request, err := r.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(120 * int64(time.Second))
	})
	if err != nil {
		log.Printf("Couldn't get a presigned request to put %v:%v. Here's why: %v\n",
			bucketName, objectKey, err)
	}
	return request.URL, err
}
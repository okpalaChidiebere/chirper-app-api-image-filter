package imagefilteraccess

import (
	"context"
	"errors"
	"fmt"
	"testing"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/okpalaChidiebere/chirper-app-api-image/v0/common"
	"github.com/stretchr/testify/assert"
)

type S3PresignerMockClient struct {
	common.PresignClientAPI
}

const (
	bucketName = "ImageFilter"
	objectKey = "barkey.jpg"
)

func initializeMockS3PresignerMockClientRepository() (PresignerRepository, error) {
	return NewPresignerRepository(&S3PresignerMockClient{}), nil
}

func (m *S3PresignerMockClient) PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error){
	result := &v4.PresignedHTTPRequest{}

	if params.Bucket == nil {
		return result, errors.New("expect bucket to not be nil")
	}
	if e, a := bucketName, *params.Bucket; e != a {
		return result, fmt.Errorf("expect %v, got %v", e, a)
	}
	if params.Key == nil {
		return result, errors.New("expect key to not be nil")
	}
	if e, a := objectKey, *params.Key; e != a {
		return result, fmt.Errorf("expect %v, got %v", e, a)
	}
	return result, nil
}


func Test_GetGetSignedUrl(t *testing.T) {
	testCases := []struct {
		name string

		bucketName string
		key string

		expectedError error
	}{
		{
			name: "Should return no error",
			bucketName: bucketName,
			key: objectKey,
			expectedError: nil,
		},
		{
			name: "Should return an error when objectKey is not provided",
			bucketName: bucketName,
			expectedError: fmt.Errorf("expect %v, got %v", objectKey, ""),
		},
		{
			name: "Should return an error when wrong bucket name is provided is not provided",
			bucketName: "ImageFilterrr",
			key: objectKey,
			expectedError: fmt.Errorf("expect %v, got %v", bucketName, "ImageFilterrr"),
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			repo, err := initializeMockS3PresignerMockClientRepository()
			if err != nil {
				t.Fatalf("error initializing repository: %s", err.Error())
			}

			_, err = repo.GetGetSignedUrl(ctx, tc.bucketName, tc.key)
			assert.Equal(t, err, tc.expectedError)
		})
	}
}
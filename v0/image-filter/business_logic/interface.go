package imagefilterservice

import (
	"context"

	"github.com/okpalaChidiebere/chirper-app-api-image/v0/common"
)

//go:generate mockgen -destination mock.go -source=interface.go -package=imagefilterservice
type Service interface {
	
	// filterImageFromURL
	// helper function to download, filter, and save the filtered image locally
	// returns the absolute path to the local image

	// INPUTS
	//    inputURL: string - a publicly accessible url to an image file or a unique key to `chirper-app-thumbnail-dev` aws bucket
	//    httpRequester: IHttpRequester - http client
	// RETURNS
	//    an absolute path to a filtered image locally saved file
	FilterImageFromURL(ctx context.Context, inputURL string,  httpRequester common.IHttpRequester) (filteredUrl string, err error)
	// UploadImageToS3Bucket
	// helper function to upload an image to `chirper-app-thumbnail-dev` aws bucket
	
	// INPUTS
	//    filePath : string - an absolute path to a filtered image locally saved file
	//    httpRequester: IHttpRequester - http client
	// RETURNS
	//    the filename of the filtered image locally saved file that was uploaded
	UploadImageToS3Bucket(ctx context.Context, filePath string,  httpRequester common.IHttpRequester) (imageKey string, err error)
	GetGetSignedUrl(ctx context.Context, bucketName, objectKey string) (url string, err error)
}
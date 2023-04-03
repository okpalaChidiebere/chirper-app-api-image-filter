package imagefilteraccess

import "context"

//go:generate mockgen -destination mock.go -source=interface.go -package=imagefilteraccess
type PresignerRepository interface {
	// GetGetSignedUrl generates an aws signed url to retrieve an item
	//  INPUT
	//	bucketName: string - the name of the aws s3 bucket to put the file
	//     objectKey: string - the filename to be put into the s3 bucket
	//  OUTPUT
	//     a url as a string
	GetGetSignedUrl(ctx context.Context, bucketName, objectKey string) (url string, err error) 
	//  GetPutSignedUrl generates an aws signed url to put an item
	//  @Params
	//	objectKey: string - the name of the aws s3 bucket to retrieve the file from
	//     objectKey: string - the filename to be retrieved from s3 bucket
	//  @Returns:
	//     a url as a string
	GetPutSignedUrl(ctx context.Context, bucketName, objectKey string) (url string, err error)
}
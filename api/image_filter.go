package api

import (
	"context"
	"errors"
	"log"

	connect "github.com/bufbuild/connect-go"
	"github.com/okpalaChidiebere/chirper-app-api-image/config"
	"github.com/okpalaChidiebere/chirper-app-api-image/v0/common"
	imagefilterservice "github.com/okpalaChidiebere/chirper-app-api-image/v0/image-filter/business_logic"
	pb "github.com/okpalaChidiebere/chirper-app-gen-protos/image_filter/v1"
	"github.com/okpalaChidiebere/chirper-app-gen-protos/image_filter/v1/image_filterv1connect"
)

var (
	mConfig    = config.NewConfig()
)

type ImageFilterServer struct {
	//image_filterv1connect.UnimplementedImagefilterServiceHandler 
	ImageFilterService imagefilterservice.Service
	HttpRequester common.IHttpRequester
}

func NewImageFilterServer(imageFilterService imagefilterservice.Service, httpRequester common.IHttpRequester ) image_filterv1connect.ImagefilterServiceHandler {
	return &ImageFilterServer{ 
		ImageFilterService: imageFilterService,
		HttpRequester: httpRequester,
	}
}

func (s *ImageFilterServer) FilterImage(ctx context.Context, req *connect.Request[pb.FilterImageRequest]) (*connect.Response[pb.FilterImageResponse], error){
	log.Println("Request headers: ", req.Header())
	// t := &model.Image {
	// 	Url: req.GetUrl(),
	// 	CreatedAt: req.GetCreatedAt(),
	// 	UpdatedAt: req.GetUpdatedAt(),
	// }

	imageUrl := req.Msg.GetImageUrl()

	//validate the imageUrl query
	if imageUrl == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument,  errors.New("image_url cannot be an empty string"))
	}

	filtered_url, err := s.ImageFilterService.FilterImageFromURL(ctx, imageUrl, s.HttpRequester)
	if err != nil {
		log.Printf("FilterImage Err: %s", err.Error())
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	objectKey, err := s.ImageFilterService.UploadImageToS3Bucket(ctx, filtered_url, s.HttpRequester) //objectKey eg: e47d4bf9-1fd1-4617-872f-ddac9f7f9084.jpg
	if err != nil {
		log.Printf("FilterImage Err: %s", err.Error())
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	//TODO: save the new image object to database

	//delete local file
	imagefilterservice.DeleteLocalFiles([]string{objectKey})

	return connect.NewResponse(&pb.FilterImageResponse{
		FilteredUrl: objectKey,
	}), nil
}

func (s *ImageFilterServer) GetGetSignedUrl(ctx context.Context, req *connect.Request[pb.GetGetSignedUrlRequest]) (*connect.Response[pb.GetGetSignedUrlResponse], error){
	objectKey := req.Msg.GetObjectKey()
	
	aws_signed_url, err := s.ImageFilterService.GetGetSignedUrl(ctx, mConfig.Dev.ImageBucket, objectKey)
	if err != nil {
		log.Printf("GetGetSignedUrl Err: %s", err.Error())
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	return connect.NewResponse(&pb.GetGetSignedUrlResponse{
		Url: aws_signed_url,
	}), nil
}
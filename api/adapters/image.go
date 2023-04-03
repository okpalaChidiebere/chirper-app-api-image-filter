package api_adapters

import (
	model "github.com/okpalaChidiebere/chirper-app-api-image/v0/image-filter/model"
	pb "github.com/okpalaChidiebere/chirper-app-gen-protos/image_filter/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ImageToProto (t *model.Image) *pb.Image{
	return &pb.Image{
		Url: t.Url,
		CreatedAt: timestamppb.New(t.CreatedAt),
		UpdatedAt: timestamppb.New(t.UpdatedAt),
	}
}

func ImagesToProto (ts []*model.Image) []*pb.Image{
	var images []*pb.Image
	for _, t := range ts {
		images = append(images, ImageToProto(t))
	}
	return images
}

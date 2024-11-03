package controller

import (
	"context"
	pb "github.com/carousell/ct-grpc-go/pkg/ct-logic-standard"
	"github.com/ct-logic-standard/internal/entity"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Controller) GetAdInfo(ctx context.Context, req *pb.GetAdInfoRequest) (res *pb.GetAdInfoResponse, err error) {
	ad, err := s.uc.GetAdByListID(ctx, req.ListId)
	return encodeResponseAdInfo(ad, err)
}

func encodeResponseAdInfo(ad entity.AdListing, err error) (*pb.GetAdInfoResponse, error) {
	if ad.AdId == 0 {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var adRes pb.AdInfo
	copier.Copy(&adRes, &ad)
	resp := pb.GetAdInfoResponse{
		Ad: &adRes,
	}
	return &resp, nil
}

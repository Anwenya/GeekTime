package grpc

import (
	"context"
	tagv1 "github.com/Anwenya/GeekTime/webook/api/proto/gen/tag/v1"
	"github.com/Anwenya/GeekTime/webook/tag/domain"
	"github.com/Anwenya/GeekTime/webook/tag/service"
	"github.com/ecodeclub/ekit/slice"
	"google.golang.org/grpc"
)

type TagServiceServer struct {
	tagv1.UnimplementedTagServiceServer
	service service.TagService
}

func NewTagServiceServer(svc service.TagService) *TagServiceServer {
	return &TagServiceServer{
		service: svc,
	}
}

func (t *TagServiceServer) Register(server grpc.ServiceRegistrar) {
	tagv1.RegisterTagServiceServer(server, t)
}

func (t *TagServiceServer) CreateTag(ctx context.Context, request *tagv1.CreateTagRequest) (*tagv1.CreateTagResponse, error) {
	id, err := t.service.CreateTag(ctx, request.Uid, request.Name)
	return &tagv1.CreateTagResponse{
		Tag: &tagv1.Tag{
			Id:   id,
			Uid:  request.Uid,
			Name: request.Name,
		},
	}, err
}

func (t *TagServiceServer) AttachTags(ctx context.Context, request *tagv1.AttachTagsRequest) (*tagv1.AttachTagsResponse, error) {
	err := t.service.AttachTags(ctx, request.Uid, request.Biz, request.BizId, request.Tids)
	return &tagv1.AttachTagsResponse{}, err
}

func (t *TagServiceServer) GetTags(ctx context.Context, request *tagv1.GetTagsRequest) (*tagv1.GetTagsResponse, error) {
	tags, err := t.service.GetTags(ctx, request.GetUid())
	if err != nil {
		return nil, err
	}
	return &tagv1.GetTagsResponse{
		Tag: slice.Map[domain.Tag](tags, func(idx int, src domain.Tag) *tagv1.Tag {
			return t.toDTO(src)
		}),
	}, nil
}

func (t *TagServiceServer) GetBizTags(ctx context.Context, req *tagv1.GetBizTagsRequest) (*tagv1.GetBizTagsResponse, error) {
	res, err := t.service.GetBizTags(ctx, req.Uid, req.Biz, req.BizId)
	if err != nil {
		return nil, err
	}
	return &tagv1.GetBizTagsResponse{
		Tags: slice.Map[domain.Tag](res, func(idx int, src domain.Tag) *tagv1.Tag {
			return t.toDTO(src)
		}),
	}, nil
}

func (t *TagServiceServer) toDTO(tag domain.Tag) *tagv1.Tag {
	return &tagv1.Tag{
		Id:   tag.Id,
		Uid:  tag.Uid,
		Name: tag.Name,
	}
}

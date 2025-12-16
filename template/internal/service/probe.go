package service

import (
	"context"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"
	pb "github.com/seanbit/kratos/template/api/web"
	"github.com/seanbit/kratos/template/internal/biz"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type ProbeService struct {
	pb.UnimplementedProbeServer
	probeBiz *biz.Probe
}

func NewProbeService(probeBiz *biz.Probe) *ProbeService {
	return &ProbeService{probeBiz: probeBiz}
}

func (s *ProbeService) Healthy(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
func (s *ProbeService) Ready(ctx context.Context, req *structpb.Struct) (*pb.ReadinessProbeResponse, error) {
	bys, err := req.MarshalJSON()
	if err != nil {
		log.Context(ctx).Errorf("marshal to json failed; detail: %+v", err)
		return nil, err
	}

	var anything map[string]interface{}
	if err = json.Unmarshal(bys, &anything); err != nil {
		log.Context(ctx).Errorf("parse argument failed: %v", err)
		return nil, err
	}

	err = s.probeBiz.Ready(ctx, anything)
	if err != nil {
		log.Context(ctx).Errorf("Readiness probe failed: %v", err)
		return nil, err
	}

	return &pb.ReadinessProbeResponse{Status: "success"}, nil
}

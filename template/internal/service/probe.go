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

// HealthLive 存活检查（Liveness）- 仅检查服务进程是否存活
func (s *ProbeService) HealthLive(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

// HealthReady 就绪检查（Readiness）- 检查服务及其依赖是否就绪
func (s *ProbeService) HealthReady(ctx context.Context, req *structpb.Struct) (*pb.ReadinessProbeResponse, error) {
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

// HealthStatus 获取详细健康状态（供监控使用）
func (s *ProbeService) HealthStatus(ctx context.Context, req *emptypb.Empty) (*pb.HealthStatusResponse, error) {
	status := s.probeBiz.CheckHealth(ctx)
	resp := &pb.HealthStatusResponse{
		Status:     status.Status,
		Components: make(map[string]*pb.HealthStatusResponse_ComponentHealth, len(status.Components)),
	}
	for idx, component := range status.Components {
		resp.Components[idx] = &pb.HealthStatusResponse_ComponentHealth{
			Status:  component.Status,
			Latency: component.Latency,
			Error:   component.Error,
		}
	}
	return resp, nil
}

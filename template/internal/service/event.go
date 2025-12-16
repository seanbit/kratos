package service

import (
	"context"

	"github.com/seanbit/kratos/template/api/event"
	"github.com/seanbit/kratos/template/internal/biz"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	EventTypeUserLogin = string(proto.MessageName(&event.UserLogin{}))
)

type EventService struct {
	event.UnimplementedEventHandlerServer
	auth *biz.Auth
}

func NewEventService(auth *biz.Auth) event.EventHandlerServer {
	return &EventService{auth: auth}
}

func (serv *EventService) HandleEvent(ctx context.Context, e *event.Event) (*emptypb.Empty, error) {
	switch e.Name {
	case EventTypeUserLogin:
		return serv.handleUserLoginEvent(ctx, e)
	}
	return &emptypb.Empty{}, nil
}

func (serv *EventService) handleUserLoginEvent(ctx context.Context, e *event.Event) (*emptypb.Empty, error) {
	message := &event.UserLogin{}
	if err := proto.Unmarshal(e.Payload, message); err != nil {
		return nil, err
	}
	err := serv.auth.SaveUserLoginLog(ctx, &biz.UserLoginLog{
		UserId:     message.UserId,
		AuthType:   message.AuthType,
		LoginIp:    message.Ip,
		LoginTime:  message.Timestamp.AsTime(),
		IssueToken: message.IssueToken,
	})
	return &emptypb.Empty{}, err
}

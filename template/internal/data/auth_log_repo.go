package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
	"github.com/seanbit/kratos/template/api/event"
	"github.com/seanbit/kratos/template/internal/biz"
	"github.com/seanbit/kratos/template/internal/data/dao"
	"github.com/seanbit/kratos/template/internal/data/model"
	"github.com/seanbit/kratos/template/internal/infra"
	"github.com/segmentio/ksuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type authLogRepo struct {
	dbProvider  infra.PostgresProvider
	rdbProvider infra.RedisProvider
	asynqClient *asynq.Client
}

func NewAuthLogRepo(dbProvider infra.PostgresProvider, rdbProvider infra.RedisProvider, asynqClient *asynq.Client) biz.IAuthLogRepo {
	return &authLogRepo{dbProvider: dbProvider, rdbProvider: rdbProvider, asynqClient: asynqClient}
}

func (repo *authLogRepo) PublishUserLoginEvent(ctx context.Context, userLoginLog *biz.UserLoginLog) error {
	message := &event.UserLogin{
		AuthType:   userLoginLog.AuthType,
		UserId:     userLoginLog.UserId,
		Ip:         userLoginLog.LoginIp,
		IssueToken: userLoginLog.IssueToken,
		Timestamp:  timestamppb.New(userLoginLog.LoginTime),
	}
	taskType := message.ProtoReflect().Descriptor().FullName()
	payload, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	task := asynq.NewTask(string(taskType), payload, asynq.TaskID(ksuid.New().String()))
	taskInfo, err := repo.asynqClient.EnqueueContext(ctx, task, asynq.MaxRetry(3), asynq.Timeout(time.Second*60))
	if err != nil {
		return err
	}
	log.Context(ctx).Infof("PublishUserLoginEvent taskInfo: %v", taskInfo)
	return nil
}

func (repo *authLogRepo) SaveUserLoginLog(ctx context.Context, userLoginLog *model.UserLoginLog) error {
	q := dao.Use(repo.dbProvider.GetDB()).UserLoginLog
	return q.WithContext(ctx).Save(userLoginLog)
}

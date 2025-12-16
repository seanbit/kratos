package data

import (
	"context"

	"github.com/pkg/errors"
	"github.com/seanbit/kratos/template/internal/biz"
	"github.com/seanbit/kratos/template/internal/data/dao"
	"github.com/seanbit/kratos/template/internal/data/model"
	"github.com/seanbit/kratos/template/internal/infra"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type authRepo struct {
	dbProvider  infra.PostgresProvider
	rdbProvider infra.RedisProvider
}

func NewAuthRepo(dbProvider infra.PostgresProvider, rdbProvider infra.RedisProvider) biz.IAuthRepo {
	return &authRepo{dbProvider: dbProvider, rdbProvider: rdbProvider}
}

func (repo *authRepo) GetUserAuthInfo(ctx context.Context, authType, authInfo string) (*model.UserAuthInfo, error) {
	userAuthInfoQ := dao.Use(repo.dbProvider.GetDB()).UserAuthInfo
	userAuthInfoDo := userAuthInfoQ.WithContext(ctx)
	record, err := userAuthInfoDo.Where(
		userAuthInfoQ.AuthType.Eq(authType),
		userAuthInfoQ.AuthInfo.Eq(authInfo),
	).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "data: get user info")
	}
	return record, err
}

func (repo *authRepo) GetUserAuthInfoByAuthType(ctx context.Context, userId, authType string) (*model.UserAuthInfo, error) {
	userAuthInfoQ := dao.Use(repo.dbProvider.GetDB()).UserAuthInfo
	userAuthInfoDo := userAuthInfoQ.WithContext(ctx)
	record, err := userAuthInfoDo.Where(
		userAuthInfoQ.UserID.Eq(userId),
		userAuthInfoQ.AuthType.Eq(authType),
	).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "data: get user info")
	}
	return record, err
}

func (repo *authRepo) SetUserAuthInfo(ctx context.Context, userAuthInfo *model.UserAuthInfo) error {
	userAuthInfoQ := dao.Use(repo.dbProvider.GetDB()).UserAuthInfo
	userAuthInfoDo := userAuthInfoQ.WithContext(ctx)
	return userAuthInfoDo.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: string(userAuthInfoQ.UserID.ColumnName())},
			{Name: string(userAuthInfoQ.AuthType.ColumnName())},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			string(userAuthInfoQ.AuthInfo.ColumnName()),
		}),
	}).Create(userAuthInfo)
}

package infra

import (
	"os"
	"sync/atomic"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
	"github.com/seanbit/kratos/template/internal/conf"
	"github.com/seanbit/kratos/webkit"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
	gormotel "gorm.io/plugin/opentelemetry/tracing"
)

type PostgresProvider interface {
	GetDB() *gorm.DB
	Close()
}
type postgresProvider struct {
	db       *gorm.DB
	cleaning *atomic.Bool
}

func NewPostgresProvider(dbCnf *conf.Data_Database) (PostgresProvider, func(), error) {
	// set logger
	gormCnf := &gorm.Config{}
	gormCnf.PrepareStmt = false
	gormCnf.SkipDefaultTransaction = true
	gormCnf.Logger = webkit.NewGormLoggerWithArgs(log.With(webkit.GetLogger(), "module", "gorm"), gormlogger.LogLevel(dbCnf.GetGormLevel()), time.Duration(dbCnf.GetSlowThreshold())*time.Millisecond)
	if dbCnf.DefaultSchema != "" {
		gormCnf.NamingStrategy = schema.NamingStrategy{TablePrefix: dbCnf.DefaultSchema + "."}
	}

	masterDB, err := gorm.Open(postgres.Open(os.Getenv("PG_DSN")), gormCnf)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	slaves := make([]gorm.Dialector, 0, len(dbCnf.GetSlavesDsn()))
	for _, slave := range dbCnf.GetSlavesDsn() {
		slaves = append(slaves, postgres.Open(slave))
	}

	// 设置主从
	if len(slaves) > 0 {
		err = masterDB.Use(dbresolver.Register(dbresolver.Config{
			Replicas: slaves,
			Policy:   dbresolver.RandomPolicy{},
		}))
		if err != nil {
			return nil, nil, errors.WithStack(err)
		}
	}

	// 设置分布式追踪和监控采集
	err = masterDB.Use(gormotel.NewPlugin())
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	err = masterDB.Use(
		dbresolver.Register(dbresolver.Config{
			// ... 同上 Sources, Replicas, Policy 配置 ...
		}).
			// 设置最大空闲连接数:cite[2]
			SetMaxIdleConns(10).
			// 设置连接的最大存活时间:cite[2]
			SetConnMaxLifetime(time.Hour).
			// 设置最大打开连接数:cite[2]
			SetMaxOpenConns(200),
	)

	sqlDB, err := masterDB.DB()
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	// configure pool parameters
	sqlDB.SetMaxOpenConns(int(dbCnf.MaxOpenConnections))
	sqlDB.SetMaxIdleConns(int(dbCnf.MaxIdleConnections))
	sqlDB.SetConnMaxLifetime(dbCnf.ConnMaxLifetime.AsDuration())
	sqlDB.SetConnMaxIdleTime(dbCnf.ConnMaxIdleTime.AsDuration())

	provider := &postgresProvider{
		db:       masterDB,
		cleaning: &atomic.Bool{},
	}
	go provider.CheckDbConnection()
	return provider, provider.Close, nil
}

func (p *postgresProvider) GetDB() *gorm.DB {
	return p.db
}

func (p *postgresProvider) Close() {
	if p.cleaning.Load() {
		return
	}
	p.cleaning.Store(true)
	sqlDB, err := p.GetDB().DB()
	if err != nil {
		log.Warnf("postgres Close error: get sql db error: %+v", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Warnf("postgres Close error: %+v", err)
		return
	}
	log.Info("postgres closed")
}

func (p *postgresProvider) CheckDbConnection() {
	for {
		p.CheckDbConnectionProcess()
		time.Sleep(time.Second * 60)
	}
}

func (p *postgresProvider) CheckDbConnectionProcess() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("CheckDbConnectionProcess: panic error; %+v", err)
		}
	}()
	sqlDB, err := p.GetDB().DB()
	if err != nil {
		log.Errorf("CheckDbConnectionProcess: GetDB error; %+v", err)
	} else {
		err = sqlDB.Ping()
		if err != nil {
			log.Errorf("CheckDbConnectionProcess: Ping error; %+v", err)
		}
	}
	stats := sqlDB.Stats()
	log.Debugf("OpenConnections: %+v", stats)
}

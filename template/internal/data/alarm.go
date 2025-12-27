package data

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/seanbit/kratos/template/internal/biz"
	"github.com/seanbit/kratos/template/internal/conf"
	"github.com/seanbit/kratos/template/internal/global"
	"github.com/seanbit/kratos/template/pkg/loghelper"
	"github.com/seanbit/kratos/webkit"
	"github.com/seanbit/kratos/webkit/thirds"
)

type AlarmMessage struct {
	Platform         string                   `json:"platform"`
	AlarmTextMessage *thirds.AlarmTextMessage `json:"alarm_text_message"`
	Retry            int                      `json:"retry"`
	MaxRetry         int                      `json:"max_retry"`
}

//go:generate mockgen -source=alarm.go -destination=./mocks/mock_alarm_message_repo.go -package=mocks
type IAlarmMessageRepo interface {
	IsIgnoreMessage(serviceName, message string) (isIgnore bool, cooldownTimes int)
	IsMessageFusing(ctx context.Context, serviceName, message string) (bool, error)
	FuseMessage(ctx context.Context, serviceName, message string, fuseDuration time.Duration) error
	IncrMessageTimes(ctx context.Context, serviceName, message string,
		cooldownTimes int, cacheIgnoreTime time.Duration) (isExceed bool, err error)
	EnqueueMessage(ctx context.Context, msg *AlarmMessage) error
	DequeueMessage(ctx context.Context) (*AlarmMessage, error)
	// 延迟队列方法
	EnqueueDelayedMessage(ctx context.Context, msg *AlarmMessage, delay time.Duration) error
	ProcessDelayedMessages(ctx context.Context) error
}

type Alarm struct {
	config      *conf.Alarm
	alarms      map[string]*thirds.Alarm // platform:alarm instance
	messageRepo IAlarmMessageRepo
	workerNum   int
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	cleaning    *atomic.Bool
}

func NewAlarm(config *conf.Alarm, messageRepo IAlarmMessageRepo) (biz.IAlarmRepo, func(), error) {
	if config.DefaultPlatform == "" || config.WebHooks[config.DefaultPlatform] == "" {
		return nil, nil, errors.New("Invalid alarm default platform configuration, please check")
	}
	if config.Concurrency == 0 {
		config.Concurrency = 3
	}
	alarms := make(map[string]*thirds.Alarm, len(config.WebHooks))
	for k, v := range config.WebHooks {
		alarms[k] = thirds.NewAlarm(global.GetServiceName(), v)
	}
	ctx, cancel := context.WithCancel(context.Background())
	alarm := &Alarm{
		config:      config,
		alarms:      alarms,
		messageRepo: messageRepo,
		workerNum:   int(config.Concurrency),
		ctx:         ctx,
		cancel:      cancel,
		cleaning:    &atomic.Bool{},
	}
	alarm.StartWorkerPool()
	return alarm, alarm.StopWorkerPool, nil
}

func (alarm *Alarm) SendBizMessage(ctx context.Context, title, info string) {
	alarm.SendMessage(ctx, "", title, info)
}

func (alarm *Alarm) SendMessage(ctx context.Context, platform, title, info string) {
	function := webkit.GetOperationFromContext(ctx)
	if function == "" {
		function = title
	}
	title = "[" + strings.ToUpper(global.GetEnv()) + "] " + title

	isExceed := false
	isIgnoreMessage, cooldownTimes := alarm.messageRepo.IsIgnoreMessage(global.GetServiceName(), info)
	if isIgnoreMessage {
		var err error
		isExceed, err = alarm.messageRepo.IncrMessageTimes(ctx, global.GetServiceName(), info, cooldownTimes, alarm.config.CacheIgnoreDuration.AsDuration())
		if err != nil {
			log.Context(ctx).Errorf("biz.Alarm.IncrMessageTimes error: %v", err)
		}
		if isExceed {
			if err = alarm.messageRepo.FuseMessage(ctx, global.GetServiceName(), info, alarm.config.CacheFuseDuration.AsDuration()); err != nil {
				log.Context(ctx).Errorf("biz.Alarm.FuseMessage error: %v", err)
			}
		}
	}

	if isFusing, err := alarm.messageRepo.IsMessageFusing(ctx, global.GetServiceName(), info); err != nil {
		log.Context(ctx).Errorf("biz.Alarm.FuseMessage error: %v", err)
	} else if isFusing {
		return
	}

	if platform == "" {
		platform = alarm.config.DefaultPlatform
	}
	if alarm.config.DryRun {
		log.Context(ctx).Debugf("dry-run alarm send text message: title:%s info: %s", fmt.Sprintf("[%s] %s", platform, title), info)
		return
	}

	alarmMessage := &AlarmMessage{
		Platform: platform,
		AlarmTextMessage: &thirds.AlarmTextMessage{
			TraceId:   webkit.GetTraceID(ctx),
			Operation: webkit.GetOperationFromContext(ctx),
			Title:     fmt.Sprintf("[%s] %s", platform, title),
			Info:      info,
		},
	}
	if err := alarm.messageRepo.EnqueueMessage(ctx, alarmMessage); err != nil {
		log.Context(ctx).Errorf("SendBizMessage:EnqueueMessage error: %v", err)
	}
}

// StartWorkerPool 启动工作池
func (alarm *Alarm) StartWorkerPool() {
	// 启动消息处理worker
	for i := 0; i < alarm.workerNum; i++ {
		alarm.wg.Add(1)
		go alarm.worker(i)
	}
	// 启动延迟队列处理器
	alarm.wg.Add(1)
	go alarm.delayedQueueProcessor()
}

// delayedQueueProcessor 延迟队列处理器
// 定期扫描延迟队列，将到期消息移入主队列
func (alarm *Alarm) delayedQueueProcessor() {
	defer alarm.wg.Done()

	log.Context(alarm.ctx).Info("Delayed queue processor started")

	ticker := time.NewTicker(time.Second) // 每秒检查一次延迟队列
	defer ticker.Stop()

	for {
		select {
		case <-alarm.ctx.Done():
			log.Context(alarm.ctx).Info("Delayed queue processor stopped")
			return
		case <-ticker.C:
			if err := alarm.messageRepo.ProcessDelayedMessages(alarm.ctx); err != nil {
				log.Context(alarm.ctx).Errorf("Process delayed messages error: %v", err)
			}
		}
	}
}

// StopWorkerPool 停止工作池
func (alarm *Alarm) StopWorkerPool() {
	if alarm.cleaning.Load() {
		return
	}
	alarm.cleaning.Store(true)
	alarm.cancel()
	log.Context(alarm.ctx).Infof("Alarm worker pool stopping...")
	alarm.wg.Wait()
	log.Context(alarm.ctx).Infof("Alarm worker pool stopped.")
}

// worker 工作协程
func (alarm *Alarm) worker(id int) {
	defer alarm.wg.Done()

	log.Context(alarm.ctx).Infof("Worker %d started", id)

	for {
		select {
		case <-alarm.ctx.Done():
			log.Context(alarm.ctx).Infof("Worker %d stopped", id)
			return
		default:
			task, err := alarm.messageRepo.DequeueMessage(alarm.ctx)
			if err != nil {
				if !errors.Is(err, redis.Nil) {
					log.Context(alarm.ctx).Errorf("Worker %d dequeue error: %v", id, err)
				}
				time.Sleep(time.Second) // 出错时等待
				continue
			}

			alarm.processMessage(task)
		}
	}
}

// processMessage 处理消息
func (alarm *Alarm) processMessage(msg *AlarmMessage) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	err := alarm.alarms[msg.Platform].SendTextMessage(ctx, msg.AlarmTextMessage)
	if err != nil {
		log.Context(ctx).Error("SendBizMessage error",
			loghelper.String("title", msg.AlarmTextMessage.Title),
			loghelper.String("msg", msg.AlarmTextMessage.Info), loghelper.FieldErr(err))
		// 重试逻辑
		if msg.Retry < msg.MaxRetry {
			msg.Retry++
			// 使用指数退避策略计算延迟时间：1s, 2s, 4s, 8s...
			retryDelay := time.Duration(1<<uint(msg.Retry-1)) * time.Second
			if retryDelay > 30*time.Second {
				retryDelay = 30 * time.Second // 最大延迟30秒
			}

			log.Context(ctx).Infof("Scheduling alarm message %s retry (attempt %d/%d) after %v",
				msg.AlarmTextMessage.TraceId, msg.Retry, msg.MaxRetry, retryDelay)

			// 使用延迟队列，避免goroutine爆炸
			if err := alarm.messageRepo.EnqueueDelayedMessage(ctx, msg, retryDelay); err != nil {
				log.Context(ctx).Errorf("Failed to enqueue delayed alarm message %s: %v",
					msg.AlarmTextMessage.TraceId, err)
			}
		} else {
			log.Context(ctx).Warnf("Alarm message %s reached max retry limit (%d), giving up",
				msg.AlarmTextMessage.TraceId, msg.MaxRetry)
		}
	} else {
		log.Context(ctx).Infof("Alarm message %s completed successfully", msg.AlarmTextMessage.TraceId)
	}
}

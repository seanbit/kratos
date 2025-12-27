package data

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"github.com/seanbit/kratos/template/internal/data/dao"
	"github.com/seanbit/kratos/template/internal/data/model"
	"github.com/seanbit/kratos/template/internal/global"
	"github.com/seanbit/kratos/template/internal/infra"
	"golang.org/x/sync/singleflight"
)

type alarmMessageRepo struct {
	dbProvider  infra.PostgresProvider
	rdbProvider infra.RedisProvider
	log         *log.Helper
	g           *singleflight.Group
}

func NewAlarmMessageRepo(dbProvider infra.PostgresProvider, rdbProvider infra.RedisProvider, logger log.Logger) IAlarmMessageRepo {
	return &alarmMessageRepo{dbProvider: dbProvider, rdbProvider: rdbProvider, log: log.NewHelper(logger), g: &singleflight.Group{}}
}

func (repo *alarmMessageRepo) IsIgnoreMessage(serviceName, message string) (isIgnore bool, cooldownTimes int) {
	ctx := context.TODO()
	rdb := repo.rdbProvider.GetRedis()
	var ignoreMessages []*model.AlarmFilterWord
	buf, _ := rdb.Get(ctx, repo.FilterWordsCacheKey(serviceName)).Bytes()
	if buf != nil && len(buf) > 0 {
		_ = json.Unmarshal(buf, &ignoreMessages)
	} else {
		ignoreMessages, _ = repo.GetAlarmFilterWordsFromDBAndCache(ctx, serviceName)
	}
	if ignoreMessages == nil || len(ignoreMessages) == 0 {
		return false, 0
	}

	for _, msg := range ignoreMessages {
		if strings.Contains(message, msg.Msg) {
			if msg.CooldownTimes <= 0 {
				msg.CooldownTimes = 50
			}
			return true, int(msg.CooldownTimes)
		}
	}
	return false, 0
}

func (repo *alarmMessageRepo) IsMessageFusing(ctx context.Context, serviceName, message string) (bool, error) {
	infoKey := repo.md5([]byte(message))
	key := repo.FilterWordsFuseKey(serviceName, infoKey)
	res, err := repo.rdbProvider.GetRedis().Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return res != "", nil
}

func (repo *alarmMessageRepo) FuseMessage(ctx context.Context, serviceName, message string, fuseDuration time.Duration) error {
	infoKey := repo.md5([]byte(message))
	key := repo.FilterWordsFuseKey(serviceName, infoKey)
	return repo.rdbProvider.GetRedis().SetNX(ctx, key, "1", fuseDuration).Err()
}

func (repo *alarmMessageRepo) IncrMessageTimes(ctx context.Context, serviceName, message string,
	cooldownTimes int, cacheIgnoreTime time.Duration) (isExceed bool, err error) {
	infoKey := repo.md5([]byte(message))
	existKey := repo.FilterWordsExistKey(serviceName, infoKey)
	timesKey := repo.FilterWordsTimesKey(serviceName, infoKey)
	rdb := repo.rdbProvider.GetRedis()
	if res, _ := rdb.Get(ctx, existKey).Result(); res != "" {
		times, _ := rdb.Get(ctx, timesKey).Int()
		if times > 0 {
			if times+1 >= cooldownTimes {
				err = rdb.Set(ctx, timesKey, 0, cacheIgnoreTime).Err()
				isExceed = true
			} else {
				err = rdb.Set(ctx, timesKey, times+1, cacheIgnoreTime).Err()
				return
			}
		} else {
			err = rdb.Set(ctx, timesKey, 1, cacheIgnoreTime).Err()
			return
		}
	} else {
		err = rdb.SetNX(ctx, existKey, "1", cacheIgnoreTime).Err()
		if err != nil {
			return
		}

		err = rdb.Set(ctx, timesKey, 1, cacheIgnoreTime).Err()
		return
	}
	return
}

func (repo *alarmMessageRepo) GetAlarmFilterWordsFromDBAndCache(ctx context.Context, serviceName string) ([]*model.AlarmFilterWord, error) {
	v, err, _ := repo.g.Do("GetAlarmFilterWordsFromDB", func() (interface{}, error) {
		ignoreMessages, err := repo.GetAlarmFilterWordsFromDB(ctx, serviceName)
		if err == nil {
			repo.rdbProvider.GetRedis().Set(ctx, repo.FilterWordsCacheKey(serviceName), ignoreMessages, time.Hour*8)
		}
		return ignoreMessages, nil
	})
	if err != nil {
		return nil, err
	}
	if v != nil {
		ignoreMessages, _ := v.([]*model.AlarmFilterWord)
		return ignoreMessages, nil
	}
	return nil, nil
}

func (repo *alarmMessageRepo) GetAlarmFilterWordsFromDB(ctx context.Context, serviceName string) ([]*model.AlarmFilterWord, error) {
	q := dao.Use(repo.dbProvider.GetDB()).AlarmFilterWord
	return q.WithContext(ctx).Where(q.Platform.Eq(serviceName)).Find()
}

func (repo *alarmMessageRepo) FilterWordsCacheKey(serviceName string) string {
	return fmt.Sprintf("%s:alarm:filter_words", serviceName)
}

func (repo *alarmMessageRepo) FilterWordsExistKey(serviceName, infoKey string) string {
	return fmt.Sprintf("%s:alarm:exist:%s", serviceName, infoKey)
}

func (repo *alarmMessageRepo) FilterWordsTimesKey(serviceName, infoKey string) string {
	return fmt.Sprintf("%s:alarm:times:%s", serviceName, infoKey)
}

func (repo *alarmMessageRepo) FilterWordsFuseKey(serviceName, infoKey string) string {
	return fmt.Sprintf("%s:alarm:fuse:%s", serviceName, infoKey)
}

func (repo *alarmMessageRepo) md5(v []byte) string {
	m := md5.New()
	m.Write(v)
	return hex.EncodeToString(m.Sum(nil))
}

// EnqueueMessage 入队
func (repo *alarmMessageRepo) EnqueueMessage(ctx context.Context, msg *AlarmMessage) error {
	taskData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal alarm message failed: %w", err)
	}

	// 使用LPUSH将任务添加到队列头部
	queueName := fmt.Sprintf("%s:alarm:message:queue", global.GetServiceName())
	err = repo.rdbProvider.GetRedis().LPush(ctx, queueName, taskData).Err()
	if err != nil {
		return fmt.Errorf("enqueue failed: %w", err)
	}

	return nil
}

// DequeueMessage 出队
func (repo *alarmMessageRepo) DequeueMessage(ctx context.Context) (*AlarmMessage, error) {
	// 使用BRPop从队列尾部获取任务，阻塞等待
	queueName := fmt.Sprintf("%s:alarm:message:queue", global.GetServiceName())
	result, err := repo.rdbProvider.GetRedis().BRPop(ctx, time.Second*10, queueName).Result()
	if err != nil {
		return nil, fmt.Errorf("dequeue alarm message failed: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid queue result")
	}

	var msg AlarmMessage
	err = json.Unmarshal([]byte(result[1]), &msg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal alarm message failed: %w", err)
	}
	return &msg, nil
}

// delayQueueKey 延迟队列的Redis key
func (repo *alarmMessageRepo) delayQueueKey() string {
	return fmt.Sprintf("%s:alarm:message:delay_queue", global.GetServiceName())
}

// EnqueueDelayedMessage 将消息加入延迟队列
// 使用Redis ZADD，score为消息应该被处理的时间戳
func (repo *alarmMessageRepo) EnqueueDelayedMessage(ctx context.Context, msg *AlarmMessage, delay time.Duration) error {
	taskData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal delayed alarm message failed: %w", err)
	}

	executeAt := float64(time.Now().Add(delay).UnixMilli())
	// 使用消息的TraceId+Retry作为member的一部分，避免重复消息被覆盖
	member := fmt.Sprintf("%s:%d:%s", msg.AlarmTextMessage.TraceId, msg.Retry, string(taskData))

	err = repo.rdbProvider.GetRedis().ZAdd(ctx, repo.delayQueueKey(), redis.Z{
		Score:  executeAt,
		Member: member,
	}).Err()
	if err != nil {
		return fmt.Errorf("enqueue delayed message failed: %w", err)
	}

	return nil
}

// ProcessDelayedMessages 处理到期的延迟消息，将其移入主队列
// 使用 ZRANGEBYSCORE 获取到期消息，然后 ZREM 删除并 LPUSH 到主队列
func (repo *alarmMessageRepo) ProcessDelayedMessages(ctx context.Context) error {
	now := float64(time.Now().UnixMilli())
	delayKey := repo.delayQueueKey()
	queueKey := fmt.Sprintf("%s:alarm:message:queue", global.GetServiceName())
	rdb := repo.rdbProvider.GetRedis()

	// 获取所有到期的消息（score <= now）
	members, err := rdb.ZRangeByScore(ctx, delayKey, &redis.ZRangeBy{
		Min:   "-inf",
		Max:   fmt.Sprintf("%f", now),
		Count: 100, // 每次最多处理100条，避免阻塞过久
	}).Result()
	if err != nil {
		return fmt.Errorf("get delayed messages failed: %w", err)
	}

	if len(members) == 0 {
		return nil
	}

	// 使用Pipeline批量处理
	pipe := rdb.Pipeline()
	for _, member := range members {
		// 解析member获取消息JSON
		// member格式: traceId:retry:jsonData
		parts := strings.SplitN(member, ":", 3)
		if len(parts) < 3 {
			repo.log.Warnf("invalid delayed message format: %s", member)
			pipe.ZRem(ctx, delayKey, member)
			continue
		}
		msgJson := parts[2]

		// 移入主队列
		pipe.LPush(ctx, queueKey, msgJson)
		// 从延迟队列删除
		pipe.ZRem(ctx, delayKey, member)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("process delayed messages pipeline failed: %w", err)
	}

	if len(members) > 0 {
		repo.log.Infof("Processed %d delayed alarm messages", len(members))
	}

	return nil
}

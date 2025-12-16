package tests

import (
	"context"
	"testing"
	"time"

	"github.com/seanbit/kratos/template/internal/data"
	"github.com/seanbit/kratos/template/internal/data/mocks"
	"github.com/seanbit/kratos/template/internal/global"
	"go.uber.org/mock/gomock"
)

func TestAlarm_SendMessageWithMock(t *testing.T) {
	type testDataItem struct {
		Platform      string
		Title         string
		Message       string
		cooldownTimes int
		timesCounter  int
		fusingEndAt   time.Time
	}
	var testData = &testDataItem{Platform: "web", Title: "Test", Message: "This is a test alarm message", cooldownTimes: 2}

	cclean := global.InitConfig(flagconfsrc, flagconf, flagsecretfile)
	defer cclean()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	messageRepo := mocks.NewMockIAlarmMessageRepo(ctrl)
	messageRepo.EXPECT().IsIgnoreMessage(gomock.Any(), gomock.Any()).Return(testData.cooldownTimes > 0, testData.cooldownTimes).AnyTimes()
	messageRepo.EXPECT().IsMessageFusing(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, serviceName, message string) (bool, error) {
			return testData.fusingEndAt.After(time.Now()), nil
		}).AnyTimes()
	messageRepo.EXPECT().FuseMessage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, serviceName, message string, fuseDuration time.Duration) error {
			testData.fusingEndAt = time.Now().Add(fuseDuration)
			return nil
		}).AnyTimes()
	messageRepo.EXPECT().IncrMessageTimes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, serviceName, message string,
			cooldownTimes int, cacheIgnoreTime time.Duration) (bool, error) {
			testData.timesCounter++
			return testData.timesCounter > testData.cooldownTimes, nil
		}).AnyTimes()

	global.GetConfig().Alarm.DryRun = true
	alarm := data.NewAlarmRepo(global.GetConfig().Alarm, messageRepo)
	for i := 0; i <= testData.cooldownTimes; i++ {
		t.Logf("send times: %d", i+1)
		alarm.SendMessage(context.TODO(), testData.Platform, testData.Title, testData.Message)
	}
}

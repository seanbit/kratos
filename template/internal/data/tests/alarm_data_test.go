package tests

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/seanbit/kratos/template/internal/data"
	"github.com/seanbit/kratos/template/internal/data/dao"
	"github.com/seanbit/kratos/template/internal/global"
	"github.com/seanbit/kratos/template/internal/infra"
	"github.com/seanbit/kratos/webkit"
)

func TestAlarmMessageRepo(t *testing.T) {
	cclean := global.InitConfig(flagconfsrc, flagconf, flagsecretfile)
	defer cclean()
	webkit.InitLogger("", "", int(global.GetConfig().LogLevel))
	dataProvider, clean, err := infra.NewDataProvider(global.GetConfig().Data)
	if err != nil {
		t.Fatal(err)
	}
	defer clean()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ignoreMessages, err := dao.Use(dataProvider.GetDB()).AlarmFilterWord.WithContext(ctx).Find()
	if err != nil {
		t.Fatal(err)
	}

	repo := data.NewAlarmMessageRepo(dataProvider, dataProvider, log.DefaultLogger)

	t.Run("TestAlarmMessageRepo_CheckAlarmFilterWords", func(t *testing.T) {
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, ignoreMessage := range ignoreMessages {
			isIgnore, cooldownTimes := repo.IsIgnoreMessage(global.GetServiceName(), ignoreMessage.Msg)
			if !isIgnore {
				t.Error("isIgnore should be true")
			}
			if cooldownTimes != int(ignoreMessage.CooldownTimes) {
				t.Error("cooldownTimes should be ", ignoreMessage.CooldownTimes)
			}
		}
	})
}

package webkit

import (
	"github.com/getsentry/sentry-go"
	"github.com/go-kratos/kratos/v2/log"
)

func InitSentry(serverName, version, env string, dsn string, withStackTrack bool) error {
	if dsn == "" {
		log.Infow("msg", "sentry disabled")
		return nil
	}
	return sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		AttachStacktrace: withStackTrack, // recommended true
		ServerName:       serverName,
		Release:          version,
		Environment:      env,
	})
}

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/go-kratos/kratos/v2/encoding/json"
	"github.com/seanbit/kratos/template/internal/global"
	"github.com/seanbit/kratos/webkit"
	"github.com/seanbit/kratos/webkit/transport/asynq"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string
	// flagconfsrc is the config source flag.
	flagconfsrc string
	// flagsecretfile is the secret file flag.
	flagsecretfile string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
	flag.StringVar(&flagconfsrc, "conf-src", "file", "config source, eg: -conf-src appconfig")
	flag.StringVar(&flagsecretfile, "secret-file", "", "secret file name, eg: -secret-file secret.yaml")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s Version: %s\n", Name, Version)
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	json.MarshalOptions = protojson.MarshalOptions{
		EmitUnpopulated: true, // Default value not ignored
		UseProtoNames:   true, // Use proto name to return http field
	}
}

func newApp(gs *grpc.Server, hs *http.Server, asynqs *asynq.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Server(
			gs,
			hs,
			asynqs,
		),
		kratos.StopTimeout(time.Second*300),
	)
}

func main() {
	flag.Parse()

	// 加载配置
	cleanconf := global.InitConfig(flagconfsrc, flagconf, flagsecretfile)
	defer cleanconf()
	bc := global.GetConfig()

	// 初始化logger
	webkit.InitLogger(Name, Version, int(bc.LogLevel))

	// 初始化Metrics
	if err := webkit.InitMetrics(Name); err != nil {
		log.Errorf("InitMetrics: %+v", err)
		panic(err)
	}

	// 初始化 sentry
	err := webkit.InitSentry(Name, Version, bc.Env.String(), bc.Sentry.GetDsn(), bc.Sentry.GetAttachStackTrace())
	if err != nil {
		log.Errorf("InitSentry: %+v", err)
		panic(err)
	}

	// 初始化分布式追踪
	if err := webkit.InitTracerProvider(bc.Tracing.GetType(), bc.Tracing.GetHost(), int(bc.Tracing.GetPort()),
		Name, Version, bc.Env.String()); err != nil {
		log.Errorf("InitTracerProvider failed: %+v", err)
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.S3, bc.GeoIp, bc.Alarm, bc.Auth, log.GetLogger())
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}

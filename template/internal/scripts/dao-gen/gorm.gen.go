// this package is used to gen some repo && po_model code
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/seanbit/kratos/template/internal/global"
	"github.com/seanbit/kratos/template/internal/infra"
	"github.com/seanbit/kratos/webkit"
	"gorm.io/gen"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// flagconf is the config flag.
	flagconf string
	// flagconfsrc is the config source flag.
	flagconfsrc string
	// flagsecretfile is the secret file flag.
	flagsecretfile string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
	flag.StringVar(&flagconfsrc, "conf-src", "file", "config source, eg: -conf-src appconfig")
	flag.StringVar(&flagsecretfile, "secret-file", "secret.yaml", "secret file name, eg: -secret-file secret.yaml")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	// pg -> update schema -> make dal
	webkit.InitLogger("gorm.gen", "v0.0.0", 0)
	// 加载配置
	cleanconf := global.InitConfig(flagconfsrc, flagconf, flagsecretfile)
	defer cleanconf()
	bc := global.GetConfig()

	// 初始化db
	dbProvider, _, err := infra.NewPostgresProvider(bc.Data.Database)
	if err != nil {
		panic(err)
	} else {
		defer dbProvider.Close()
	}

	db := dbProvider.GetDB()

	g := gen.NewGenerator(gen.Config{
		OutPath: "internal/data/dao",
		Mode:    gen.WithQueryInterface,
		//FieldNullable:     false,
		//FieldCoverable:    false,
		//FieldSignable:     false,
		//FieldWithIndexTag: false,
		FieldWithTypeTag: true,
	})

	g.UseDB(db)
	generateModels(g)
	g.Execute()
}

func generateModels(g *gen.Generator) {
	exportIndexBackendModels(g)
}

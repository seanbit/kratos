package tests

import (
	"flag"
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/encoding/json"
	"google.golang.org/protobuf/encoding/protojson"
)

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
	flag.StringVar(&flagconf, "conf", "../../../configs", "config path, eg: -conf config.yaml")
	flag.StringVar(&flagconfsrc, "conf-src", "file", "config source, eg: -conf-src appconfig")
	flag.StringVar(&flagsecretfile, "secret-file", "secret.yaml", "secret file name, eg: -secret-file secret.yaml")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	json.MarshalOptions = protojson.MarshalOptions{
		EmitUnpopulated: true, // Default value not ignored
		UseProtoNames:   true, // Use proto name to return http field
	}
	if os.Getenv("TEST_CONFIG_FILE") != "" {
		flagconf = os.Getenv("TEST_CONFIG_FILE")
	}
	if os.Getenv("TEST_SECRET_FILE") != "" {
		flagsecretfile = os.Getenv("TEST_SECRET_FILE")
	}
}

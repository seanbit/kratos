package example

import (
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/seanbit/kratos/template/cmd/job/jobs"
	"github.com/spf13/cobra"
)

var (
	name string
)

type exampleJob struct{}

func (j *exampleJob) Name() string  { return "example" }
func (j *exampleJob) Short() string { return "An example job" }
func (j *exampleJob) Long() string {
	return "An example job demonstrating the job runner pattern"
}

func (j *exampleJob) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&name, "name", "n", "World", "Name to greet")
}

func (j *exampleJob) Run(app *jobs.App, cmd *cobra.Command, args []string) error {
	log.NewHelper(app.Logger).Infof("example job started, name=%s", name)
	fmt.Printf("Hello, %s!\n", name)
	return nil
}

func init() {
	jobs.Register(&exampleJob{})
}

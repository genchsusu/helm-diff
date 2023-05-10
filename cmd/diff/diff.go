package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

var settings = cli.New()

func warning(format string, v ...interface{}) {
	format = fmt.Sprintf("WARNING: %s\n", format)
	fmt.Fprintf(os.Stderr, format, v...)
}

func main() {

	actionConfig := new(action.Configuration)
	cmd, err := newRootCmd(actionConfig, os.Stdout, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	cobra.OnInitialize(func() {
		helmDriver := os.Getenv("HELM_DRIVER")
		if err := actionConfig.Init(settings.RESTClientGetter(), namespace, helmDriver, nil); err != nil {
			log.Fatal(err)
		}
	})

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

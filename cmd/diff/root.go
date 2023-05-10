package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
)

var namespace string

func newRootCmd(actionConfig *action.Configuration, out io.Writer, args []string) (*cobra.Command, error) {
	clientGet := action.NewGet(actionConfig)
	clientGet.Version = 0

	clientInstall := action.NewInstall(actionConfig)
	valueOpts := &values.Options{}

	cmd := &cobra.Command{
		Use:          "diff [RELEASE] [CHART] [flags]",
		Short:        "Show manifest differences",
		Long:         "Show manifest differences",
		SilenceUsage: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 || len(args) > 2 {
				cmd.Help()
				os.Exit(1)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := clientGet.Run(args[0])
			if err != nil {
				return err
			}
			lastReleaseManifest := strings.TrimSpace(res.Manifest)
			// fmt.Fprintln(out, lastReleaseManifest)
			lastReleaseManifest = removeSecrets(lastReleaseManifest)

			clientInstall.DryRun = true
			clientInstall.Replace = true
			clientInstall.ClientOnly = true
			clientInstall.DisableHooks = true
			clientInstall.IncludeCRDs = false

			rel, err := runInstall(args, clientInstall, valueOpts, out)

			if err != nil && !settings.Debug {
				if rel != nil {
					return fmt.Errorf("%w\n\nUse --debug flag to render out invalid YAML", err)
				}
				return err
			}
			if rel != nil {
				var manifests bytes.Buffer
				fmt.Fprintln(&manifests, strings.TrimSpace(rel.Manifest))

				chartTemplate := manifests.String()
				// fmt.Fprintln(out, chartTemplate)
				chartTemplate = removeSecrets(chartTemplate)

				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(lastReleaseManifest, chartTemplate, false)
				fmt.Fprintln(out, dmp.DiffPrettyText(diffs))
			}

			return nil
		},
	}

	f := cmd.Flags()
	addInstallFlags(cmd, f, valueOpts, settings)

	return cmd, nil
}

func runInstall(args []string, client *action.Install, valueOpts *values.Options, out io.Writer) (*release.Release, error) {
	name, chart, err := client.NameAndChart(args)
	if err != nil {
		return nil, err
	}
	client.ReleaseName = name

	cp, err := client.ChartPathOptions.LocateChart(chart, settings)
	if err != nil {
		return nil, err
	}

	p := getter.All(settings)
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		return nil, err
	}

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	if err := checkIfInstallable(chartRequested); err != nil {
		return nil, err
	}

	if chartRequested.Metadata.Deprecated {
		warning("This chart is deprecated")
	}

	client.Namespace = settings.Namespace()

	// Create context and prepare the handle of SIGTERM
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	cSignal := make(chan os.Signal, 2)
	signal.Notify(cSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-cSignal
		fmt.Fprintf(out, "Release %s has been cancelled.\n", args[0])
		cancel()
	}()

	return client.RunWithContext(ctx, chartRequested, vals)
}

// checkIfInstallable validates if a chart can be installed
//
// Application chart type is only installable
func checkIfInstallable(ch *chart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	return errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}

func addInstallFlags(cmd *cobra.Command, f *pflag.FlagSet, valueOpts *values.Options, settings *cli.EnvSettings) {
	addValueOptionsFlags(f, valueOpts)
	f.StringVarP(&namespace, "namespace", "n", settings.Namespace(), "namespace scope for this request")
}

func addValueOptionsFlags(f *pflag.FlagSet, v *values.Options) {
	f.StringSliceVarP(&v.ValueFiles, "values", "f", []string{}, "specify values in a YAML file or a URL (can specify multiple)")
	f.StringArrayVar(&v.Values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&v.StringValues, "set-string", []string{}, "set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&v.FileValues, "set-file", []string{}, "set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)")
}

type manifestYaml struct {
	ApiVersion string
	Kind       string
}

func removeSecrets(content string) string {
	var wantedList []string
	contentList := strings.Split(content, "---\n")

	for _, r := range contentList {
		var data manifestYaml
		_ = yaml.Unmarshal([]byte(r), &data)
		if data.Kind != "Secret" {
			wantedList = append(wantedList, r)
		}
	}

	return strings.Join(wantedList, "---\n")
}

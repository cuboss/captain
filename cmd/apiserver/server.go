package main

import (
	"captain/pkg/version"
	"context"
	"fmt"
	"github.com/spf13/cobra"
	cliflag "k8s.io/component-base/cli/flag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/component-base/term"
	"k8s.io/klog"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	apiserverconfig "captain/pkg/apiserver/config"
	"captain/cmd/apiserver/options"
)

func NewAPIServerCommand() *cobra.Command {
	s := options.NewServerRunOptions()

	// Load configuration from file
	conf, err := apiserverconfig.TryLoadFromDisk()
	if err == nil {
		s = &options.ServerRunOptions{
			GenericServerRunOptions: s.GenericServerRunOptions,
			Config:                  conf,
		}
	} else {
		klog.Fatal("Failed to load configuration from disk", err)
	}

	cmd := &cobra.Command{
		Use: "captain-apiserver",
		Long: `The Captain API server validates and configures data for the API objects. 
The API Server services REST operations and provides the frontend to the
cluster's shared state through which all other components interact.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if errs := s.Validate(); len(errs) != 0 {
				return utilerrors.NewAggregate(errs)
			}
			return Run(s, apiserverconfig.WatchConfigChange(), signals.SetupSignalHandler())
		},
		SilenceUsage: true,
	}

	fs := cmd.Flags()
	namedFlagSets := s.Flags()
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of Captain apiserver",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(version.Get())
		},
	}

	cmd.AddCommand(versionCmd)

	return cmd
}

func Run(s *options.ServerRunOptions, configCh <-chan apiserverconfig.Config, ctx context.Context) error {
	ictx, cancelFunc := context.WithCancel(context.TODO())
	errCh := make(chan error)
	defer close(errCh)
	go func() {
		if err := run(s, ictx); err != nil {
			errCh <- err
		}
	}()

	// The ctx (signals.SetupSignalHandler()) is to control the entire program life cycle,
	// The ictx(internal context)  is created here to control the life cycle of the ks-apiserver(http server, sharedInformer etc.)
	// when config change, stop server and renew context, start new server
	for {
		select {
		case <-ctx.Done():
			cancelFunc()
			return nil
		case cfg := <-configCh:
			cancelFunc()
			s.Config = &cfg
			ictx, cancelFunc = context.WithCancel(context.TODO())
			go func() {
				if err := run(s, ictx); err != nil {
					errCh <- err
				}
			}()
		case err := <-errCh:
			cancelFunc()
			return err
		}
	}
}

func run(s *options.ServerRunOptions, ctx context.Context) error {
	apiserver, err := s.NewAPIServer(ctx.Done())
	if err != nil {
		return err
	}

	err = apiserver.PrepareRun(ctx.Done())
	if err != nil {
		return err
	}

	err = apiserver.Run(ctx)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
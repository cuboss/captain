package app

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/term"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"captain/cmd/controller-manager/app/options"
	"captain/pkg/apis"
	controllerconfig "captain/pkg/server/config"
	"captain/pkg/server/informers"
	"captain/pkg/simple/client/k8s"
	"captain/pkg/version"
)

func NewControllerManagerCommand() *cobra.Command {
	s := options.NewCaptainControllerManagerOptions()
	conf, err := controllerconfig.TryLoadFromDisk()
	if err == nil {
		// make sure LeaderElection is not nil
		s = &options.CaptainControllerManagerOptions{
			KubernetesOptions:   conf.KubernetesOptions,
			MultiClusterOptions: conf.MultiClusterOptions,
			LeaderElection:      s.LeaderElection,
			LeaderElect:         s.LeaderElect,
			WebhookCertDir:      s.WebhookCertDir,
		}
	} else {
		klog.Fatal("Failed to load configuration from disk", err)
	}

	cmd := &cobra.Command{
		Use:  "controller-manager",
		Long: `Captain controller manager is a daemon that`,
		Run: func(cmd *cobra.Command, args []string) {
			if errs := s.Validate(); len(errs) != 0 {
				klog.Error(utilerrors.NewAggregate(errs))
				os.Exit(1)
			}

			if err = run(s, signals.SetupSignalHandler()); err != nil {
				klog.Error(err)
				os.Exit(1)
			}
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
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of Captain controller-manager",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(version.Get())
		},
	}

	cmd.AddCommand(versionCmd)

	return cmd
}

func run(s *options.CaptainControllerManagerOptions, ctx context.Context) error {

	kubernetesClient, err := k8s.NewKubernetesClient(s.KubernetesOptions)
	if err != nil {
		klog.Errorf("Failed to create kubernetes clientset %v", err)
		return err
	}

	informerFactory := informers.NewInformerFactories(
		kubernetesClient.Kubernetes(),
		kubernetesClient.Crd(),
	)

	mgrOptions := manager.Options{
		CertDir: s.WebhookCertDir,
		Port:    8443,
	}

	if s.LeaderElect {
		mgrOptions = manager.Options{
			CertDir:                 s.WebhookCertDir,
			Port:                    8443,
			LeaderElection:          s.LeaderElect,
			LeaderElectionNamespace: "captain-system",
			LeaderElectionID:        "captain-controller-manager-leader-election",
			LeaseDuration:           &s.LeaderElection.LeaseDuration,
			RetryPeriod:             &s.LeaderElection.RetryPeriod,
			RenewDeadline:           &s.LeaderElection.RenewDeadline,
		}
	}

	klog.V(0).Info("setting up manager")

	// TODO
	//ctrl.SetLogger(klogr.New().V(0))

	// Use 8443 instead of 443 cause we need root permission to bind port 443
	mgr, err := manager.New(kubernetesClient.Config(), mgrOptions)
	if err != nil {
		klog.Fatalf("unable to set up overall controller manager: %v", err)
	}

	if err = apis.AddToScheme(mgr.GetScheme()); err != nil {
		klog.Fatalf("unable add APIs to scheme: %v", err)
	}

	// register common meta types into schemas.
	metav1.AddToGroupVersion(mgr.GetScheme(), metav1.SchemeGroupVersion)

	if err = addControllers(mgr,
		kubernetesClient,
		informerFactory,
		s.KubernetesOptions,
		s.MultiClusterOptions,
		ctx.Done()); err != nil {
		klog.Fatalf("unable to register controllers to the manager: %v", err)
	}

	// Start cache data after all informer is registered
	klog.V(0).Info("Starting cache resource from apiserver...")
	informerFactory.Start(ctx.Done())

	klog.V(0).Info("Starting the controllers.")
	if err = mgr.Start(ctx); err != nil {
		klog.Fatalf("unable to run the manager: %v", err)
	}

	return nil
}

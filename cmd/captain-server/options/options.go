package options

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"k8s.io/klog"

	"captain/pkg/informers"
	"captain/pkg/server"
	captainserverconfig "captain/pkg/server/config"
	"captain/pkg/simple/client/k8s"
	"captain/pkg/simple/client/monitoring/prometheus"
	genericoptions "captain/pkg/simple/server/options"

	cliflag "k8s.io/component-base/cli/flag"
)

type ServerRunOptions struct {
	ConfigFile              string
	GenericServerRunOptions *genericoptions.ServerRunOptions
	*captainserverconfig.Config

	//
	DebugMode bool
}

func NewServerRunOptions() *ServerRunOptions {
	s := &ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
		Config:                  captainserverconfig.New(),
	}

	return s
}

func (s *ServerRunOptions) Flags() (fss cliflag.NamedFlagSets) {
	fs := fss.FlagSet("generic")
	fs.BoolVar(&s.DebugMode, "debug", false, "Don't enable this if you don't know what it means.")
	s.GenericServerRunOptions.AddFlags(fs, s.GenericServerRunOptions)
	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"), s.KubernetesOptions)
	s.RedisOptions.AddFlags(fss.FlagSet("redis"), s.RedisOptions)
	s.MonitoringOptions.AddFlags(fss.FlagSet("monitoring"), s.MonitoringOptions)

	fs = fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		fs.AddGoFlag(fl)
	})

	return fss
}

func (s *ServerRunOptions) NewAPIServer(stopCh <-chan struct{}) (*server.CaptainAPIServer, error) {
	apiServer := &server.CaptainAPIServer{
		Config: s.Config,
	}

	kubernetesClient, err := k8s.NewKubernetesClient(s.KubernetesOptions)
	if err != nil {
		return nil, err
	}
	apiServer.KubernetesClient = kubernetesClient

	informerFactory := informers.NewInformerFactories(kubernetesClient.Kubernetes(), kubernetesClient.Crd())
	apiServer.InformerFactory = informerFactory

	captainServer := &http.Server{
		Addr: fmt.Sprintf(":%d", s.GenericServerRunOptions.InsecurePort),
	}
	apiServer.Server = captainServer

	if s.MonitoringOptions != nil && len(s.MonitoringOptions.Endpoint) > 0 {
		monitoringClient, err := prometheus.NewPrometheus(s.MonitoringOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to prometheus, please check prometheus status, error: %v", err)
		}
		apiServer.MonitoringClient = monitoringClient
	} else {
		klog.Warning("moinitoring service address in configuration MUST not be empty, please check configmap/captain-config in captain-system namespace")
	}

	return apiServer, nil
}

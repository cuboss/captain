package prometheus

import "github.com/spf13/pflag"

type Options struct {
	Endpoint string         `json:"endpoint,omitempty" yaml:"endpoint"`
	Auth     PrometheusAuth `json:"auth,omitempty" yaml:"auth"`
}

type PrometheusAuth struct {
	Basic struct {
		Username string `json:"username,omitempty" yaml:"username"`
		Password string `json:"password,omitempty" yaml:"password"`
	} `json:"basic,omitempty" yaml:"basic"`
}

func NewPrometheusOptions() *Options {
	return &Options{
		Endpoint: "",
		Auth:     PrometheusAuth{},
	}
}

func (s *Options) Validate() []error {
	var errs []error
	return errs
}

func (s *Options) ApplyTo(options *Options) {
	if s.Endpoint != "" {
		options.Endpoint = s.Endpoint
	}
}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.StringVar(&s.Endpoint, "prometheus-endpoint", c.Endpoint, ""+
		"Prometheus service endpoint which stores KubeSphere monitoring data, if left "+
		"blank, will use builtin metrics-server as data source.")
	fs.StringVar(&s.Auth.Basic.Username, "prometheus-username", c.Auth.Basic.Username, "Prometheus service username.")
	fs.StringVar(&s.Auth.Basic.Password, "prometheus-password", c.Auth.Basic.Password, "Prometheus service pasword.")
}

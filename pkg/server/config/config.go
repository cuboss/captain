package config

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"

	"captain/pkg/constants"
	"captain/pkg/simple/client/cache"
	"captain/pkg/simple/client/helm"
	"captain/pkg/simple/client/k8s"
	"captain/pkg/simple/client/monitoring/prometheus"
	"captain/pkg/simple/client/multicluster"
)

// Package config saves configuration for running KubeSphere components
//
// Config can be configured from command line flags and configuration file.
// Command line flags hold higher priority than configuration file. But if
// component Endpoint/Host/APIServer was left empty, all of that component
// command line flags will be ignored, use configuration file instead.
// For example, we have configuration file
//
// mysql:
//   host: mysql.kubesphere-system.svc
//   username: root
//   password: password
//
// At the same time, have command line flags like following:
//
// --mysql-host mysql.openpitrix-system.svc --mysql-username king --mysql-password 1234
//

var (
	// singleton instance of config package
	_config = defaultConfig()
)

const (
	// DefaultConfigurationName is the default name of configuration
	defaultConfigurationName = "captain"

	// DefaultConfigurationPath the default location of the configuration file
	defaultConfigurationPath = "/etc/captain"
)

type config struct {
	cfg         *Config
	cfgChangeCh chan Config
	watchOnce   sync.Once
	loadOnce    sync.Once
}

func (c *config) watchConfig() <-chan Config {
	c.watchOnce.Do(func() {
		viper.WatchConfig()
		viper.OnConfigChange(func(in fsnotify.Event) {
			cfg := New()
			if err := viper.Unmarshal(cfg); err != nil {
				klog.Warning("config reload error", err)
			} else {
				c.cfgChangeCh <- *cfg
			}
		})
	})
	return c.cfgChangeCh
}

func (c *config) loadFromDisk() (*Config, error) {
	var err error
	c.loadOnce.Do(func() {
		if err = viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				err = fmt.Errorf("error parsing configuration file %s", err)
			}
		}
		err = viper.Unmarshal(c.cfg)
	})
	return c.cfg, err
}

func defaultConfig() *config {
	viper.SetConfigName(defaultConfigurationName)
	viper.AddConfigPath(defaultConfigurationPath)

	// Load from current working directory, only used for debugging
	viper.AddConfigPath(".")

	// Load from Environment variables
	viper.SetEnvPrefix("captain")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return &config{
		cfg:         New(),
		cfgChangeCh: make(chan Config),
		watchOnce:   sync.Once{},
		loadOnce:    sync.Once{},
	}
}

// Config defines everything needed for captain-server to deal with external services
type Config struct {
	KubernetesOptions   *k8s.KubernetesOptions `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty" mapstructure:"kubernetes"`
	RedisOptions        *cache.Options         `json:"redis,omitempty" yaml:"redis,omitempty" mapstructure:"redis"`
	MultiClusterOptions *multicluster.Options  `json:"multicluster,omitempty" yaml:"multicluster,omitempty" mapstructure:"multicluster"`
	MonitoringOptions   *prometheus.Options    `json:"monitoring,omitempty" yaml:"monitoring,omitempty" mapstructure:"monitoring"`
	ComponentOptions    *helm.Options          `json:"component,omitempty" yaml:"component,omitempty" mapstructure:"component"`
}

// newConfig creates a default non-empty Config
func New() *Config {
	return &Config{
		KubernetesOptions:   k8s.NewKubernetesOptions(),
		RedisOptions:        cache.NewRedisOptions(),
		MultiClusterOptions: multicluster.NewOptions(),
		MonitoringOptions:   prometheus.NewPrometheusOptions(),
		ComponentOptions:    helm.NewOptions(),
	}
}

// TryLoadFromDisk loads configuration from default location after server startup
// return nil error if configuration file not exists
func TryLoadFromDisk() (*Config, error) {
	return _config.loadFromDisk()
}

// WatchConfigChange return config change channel
func WatchConfigChange() <-chan Config {
	return _config.watchConfig()
}

// convertToMap simply converts config to map[string]bool
// to hide sensitive information
func (conf *Config) ToMap() map[string]bool {
	conf.stripEmptyOptions()
	result := make(map[string]bool)

	if conf == nil {
		return result
	}

	c := reflect.Indirect(reflect.ValueOf(conf))

	for i := 0; i < c.NumField(); i++ {
		name := strings.Split(c.Type().Field(i).Tag.Get("json"), ",")[0]
		if strings.HasPrefix(name, "-") {
			continue
		}

		if c.Field(i).IsNil() {
			result[name] = false
		} else {
			result[name] = true
		}
	}

	return result
}

// Remove invalid options before serializing to json or yaml
func (conf *Config) stripEmptyOptions() {

	if conf.RedisOptions != nil && conf.RedisOptions.Host == "" {
		conf.RedisOptions = nil
	}

	if conf.MonitoringOptions != nil && conf.MonitoringOptions.Endpoint == "" {
		conf.MonitoringOptions = nil
	}

}

// GetFromConfigMap returns KubeSphere ruuning config by the given ConfigMap.
func GetFromConfigMap(cm *corev1.ConfigMap) (*Config, error) {
	c := &Config{}
	value, ok := cm.Data[constants.CaptainConfigMapDataKey]
	if !ok {
		return nil, fmt.Errorf("failed to get configmap captain.yaml value")
	}

	if err := yaml.Unmarshal([]byte(value), c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal value from configmap. err: %s", err)
	}
	return c, nil
}

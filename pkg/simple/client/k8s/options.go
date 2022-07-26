package k8s

import (
	"os"
	"os/user"
	"path"

	"k8s.io/client-go/util/homedir"

	"github.com/spf13/pflag"

	"captain/pkg/utils/reflectutils"
)

type KubernetesOptions struct {
	// kubeconfig path, if not specified, will use
	// in cluster way to create clientset
	KubeConfig string `json:"kubeconfig" yaml:"kubeconfig"`

	// kubernetes captain-server public address, used to generate kubeconfig
	// for downloading, default to host defined in kubeconfig
	// +optional
	Master string `json:"master,omitempty" yaml:"master,omitempty"`

	// kubernetes clientset qps
	// +optional
	QPS float32 `json:"qps,omitempty" yaml:"qps,omitempty"`

	// kubernetes clientset burst
	// +optional
	Burst int `json:"burst,omitempty" yaml:"burst,omitempty"`
}

// NewKubernetesOptions returns a `zero` instance
func NewKubernetesOptions() (option *KubernetesOptions) {
	option = &KubernetesOptions{
		QPS:   1e6,
		Burst: 1e6,
	}

	// make it be easier for those who wants to run api-server locally
	homePath := homedir.HomeDir()
	if homePath == "" {
		// try os/user.HomeDir when $HOME is unset.
		if u, err := user.Current(); err == nil {
			homePath = u.HomeDir
		}
	}

	userHomeConfig := path.Join(homePath, ".kube/config")
	if _, err := os.Stat(userHomeConfig); err == nil {
		option.KubeConfig = userHomeConfig
	}
	return
}

func (k *KubernetesOptions) Validate() []error {
	errors := []error{}

	if k.KubeConfig != "" {
		if _, err := os.Stat(k.KubeConfig); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (k *KubernetesOptions) ApplyTo(options *KubernetesOptions) {
	reflectutils.Override(options, k)
}

func (k *KubernetesOptions) AddFlags(fs *pflag.FlagSet, c *KubernetesOptions) {
	fs.StringVar(&k.KubeConfig, "kubeconfig", c.KubeConfig, ""+
		"Path for kubernetes kubeconfig file, if left blank, will use "+
		"in cluster way.")

	fs.StringVar(&k.Master, "master", c.Master, ""+
		"Used to generate kubeconfig for downloading, if not specified, will use host in kubeconfig.")
}

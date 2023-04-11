package helm

import "github.com/spf13/pflag"

type Options struct {
	Name     string `json:"name" yaml:"name"`
	URL      string `json:"url" yaml:"url"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

// NewOptions returns a default nil options
func NewOptions() *Options {
	return &Options{}
}

func (o *Options) Validate() []error {
	return nil
}

func (o *Options) AddFlags(fs *pflag.FlagSet, s *Options) {
	fs.StringVar(&o.Name, "helm-repo-name", s.Name, "the name of helm repo")
	fs.StringVar(&o.URL, "helm-repo-url", s.URL, "the url of helm repo")
	fs.StringVar(&o.Username, "helm-repo-username", s.Name, "the username of helm repo")
	fs.StringVar(&o.Password, "helm-repo-password", s.URL, "the password of helm repo")
}

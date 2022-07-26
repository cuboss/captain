
package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"


	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"



	"captain/pkg/simple/client/cache"

)

func newTestConfig() (*Config, error) {
	var conf = &Config{

		RedisOptions: &cache.Options{
			Host:     "localhost",
			Port:     6379,
			Password: "Acd13G",
			DB:       0,
		},

	}
	return conf, nil
}

func saveTestConfig(t *testing.T, conf *Config) {
	content, err := yaml.Marshal(conf)
	if err != nil {
		t.Fatalf("error marshal config. %v", err)
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s.yaml", defaultConfigurationName), content, 0640)
	if err != nil {
		t.Fatalf("error write configuration file, %v", err)
	}
}

func cleanTestConfig(t *testing.T) {
	file := fmt.Sprintf("%s.yaml", defaultConfigurationName)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Log("file not exists, skipping")
		return
	}

	err := os.Remove(file)
	if err != nil {
		t.Fatalf("remove %s file failed", file)
	}

}

func TestGet(t *testing.T) {
	conf, err := newTestConfig()
	if err != nil {
		t.Fatal(err)
	}
	saveTestConfig(t, conf)
	defer cleanTestConfig(t)

	conf.RedisOptions.Password = "P@88w0rd"
	os.Setenv("KUBESPHERE_REDIS_PASSWORD", "P@88w0rd")

	conf2, err := TryLoadFromDisk()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(conf, conf2); diff != "" {
		t.Fatal(diff)
	}
}

func TestStripEmptyOptions(t *testing.T) {
	var config Config

	config.RedisOptions = &cache.Options{Host: ""}


	config.stripEmptyOptions()

	if config.RedisOptions != nil {
		t.Fatal("config stripEmptyOptions failed")
	}
}

package main

import (
	"github.com/go-yaml/yaml"
	"github.com/imdario/mergo"
	"io/ioutil"
	"log"
	"os"
)

const (
	DEFAULT_CONFIG_PATH1 = "~/.config/geoipcli.yaml"
	DEFAULT_CONFIG_PATH2 = "~/.geoipcli.yaml"
)

var (
	// This is the master configuration data.
	Config CLIConfig
)

type CLIConfig struct {
	Paths struct {
		Country        string `yaml:"country"`
		City           string `yaml:"city"`
		ASN            string `yaml:"asn"`
		ISP            string `yaml:"isp"`
		Domain         string `yaml:"domain"`
		ConnectionType string `yaml:"connection_type"`
		AnonymousIP    string `yaml:"anonymousip"`
		Enterprise     string `yaml:"enterprise"`
	} `yaml:"paths"`
	Output struct {
		Format        string   `yaml:"format"`
		Columns       []string `yaml:"columns"`
		SkipInvalidIP bool     `yaml:"skip_invalid_ip"`
	} `yaml:"output"`
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func LoadDefaultConfigs() {
	for _, path := range []string{DEFAULT_CONFIG_PATH1, DEFAULT_CONFIG_PATH2} {
		if !fileExists(path) {
			if Debug {
				log.Println("[-] Skip default config (not found):", path)
			}
			continue
		}

		if Debug {
			log.Println("[+] Load default config:", path)
		}
		LoadConfig(path)
	}
}

// Load data from file,
// parse the data as YAML format,
// merge the structured data to the master config data.
func LoadConfig(filename string) {
	if Debug {
		log.Println("[+] Load config:", filename)
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalln("[-] Failed to read config file:", err)
	}

	c := CLIConfig{}
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		log.Fatalln("[-] Failed to parse config file:", err)
	}

	if err := mergo.Merge(&Config, c); err != nil {
		log.Fatalln("[-] Failed to merge config data:", err)
	}
}

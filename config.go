package main

import (
	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Paths struct {
		Country        string `yaml:"country"`
		City           string `yaml:"city"`
		ASN            string `yaml:"asn"`
		ISP            string `yaml:"isp"`
		Domain         string `yaml:"domain"`
		ConnectionType string `yaml:"connection_type"`
		AnonymousIP    string `yaml:"anonymousip"`
	} `yaml:"paths"`
	Output struct {
		// Output format (csv, tsv)
		Format            string   `yaml:"format"`
		// List of output columns.
		Columns           []string `yaml:"columns"`
		// Flag if commas are escaped.
		EscapeComma       bool     `yaml:"escape_comma"`
		// Flag if double quotes are escaped.
		EscapeDoubleQuote bool     `yaml:"escape_double_quote"`
		// Flag if invalid IP addresses are skipped.
		SkipInvalidIP     bool     `yaml:"skip_invalid_ip"`
	} `yaml:"output"`
}

// Load configuration data from file and merge the data to the config data.
func LoadConfig(config *Config, filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	c := Config{}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return err
	}

	err = mergo.Merge(config, c)
	if err != nil {
		return err
	}

	return nil
}

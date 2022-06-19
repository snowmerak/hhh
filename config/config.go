package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Listening is the configuration for the listening addressses.
type Listening struct {
	Address            string `yaml:"address"`
	CertificatePemFile string `yaml:"certificate_pem_file"`
	CertificateKeyFile string `yaml:"certificate_key_file"`
}

// Config is the configuration for the application
type Config struct {
	Listenings            []Listening `yaml:"listenings"`
	ReverseProxyAddresses []string    `yaml:"reverse_proxy_addresses"`
	LimitPerMillisecond   float64     `yaml:"limit_per_millisecond"`
	MaxTryCount           int         `yaml:"max_try_count"`
}

func InitAndCreate(fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("InitAndCreate: os.Create: %w", err)
	}
	defer f.Close()
	cf := Config{
		Listenings: []Listening{
			{
				Address:            ":8080",
				CertificatePemFile: "cert.pem",
				CertificateKeyFile: "key.pem",
			},
		},
		LimitPerMillisecond: 1000,
	}
	encoder := yaml.NewEncoder(f)
	if err := encoder.Encode(&cf); err != nil {
		return fmt.Errorf("InitAndCreate: encoder.Encode: %w", err)
	}
	return nil
}

func ReadAndParse(fileName string, config *Config) error {
	f, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("ReadAndParse: os.Open: %w", err)
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&config); err != nil {
		return fmt.Errorf("ReadAndParse: decoder.Decode: %w", err)
	}
	return nil
}

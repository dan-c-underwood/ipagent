package ipagent

import (
	"github.com/spf13/viper"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// RecordType represents the type for DNS records.
type RecordType string

const (
	A RecordType = "A"
	AAAA RecordType = "AAAA"
	CNAME RecordType = "CNAME"
	MX RecordType = "MX"
)

// IsValid checks whether a RecordType string is a valid DNS record type (e.g. A, CNAME).
func (rt RecordType) IsValid() bool {
	switch rt {
	case A, AAAA, CNAME, MX:
		return true
	default:
		return false
	}
}

// Config holds the required configuration for ipagent.
type Config struct {
	Logging bool `mapstructure:"logging"`
	Cloudflare CloudflareConfig `mapstructure:"cloudflare"`
	Domain Domain `mapstructure:"domain"`
	SubDomains []Domain `mapstructure:"sub_domains"`
}

// GetDomainList generates a list of both the domain and subdomains contained within a Config value.
func (c Config) GetDomainList() []Domain {
	dList := make([]Domain, len(c.SubDomains) + 1)
	dList[0] = c.Domain

	for i, d := range c.SubDomains {
		d.Name = d.Name + "." + c.Domain.Name
		dList[i+1] = d
	}

	return dList
}

// CloudflareConfig holds the configuration values required to use the Cloudflare API.
type CloudflareConfig struct {
	ZoneID string `mapstructure:"zone_id"`
	APIKey string `mapstructure:"api_key"`
	APIEmail string `mapstructure:"api_email"`
}

// Domain contains the values that represent a domain in a domain record on Cloudflare.
type Domain struct {
	Name string `mapstructure:"name"`
	Proxy bool `mapstructure:"proxy"`
	Type RecordType `mapstructure:"type"`
}

// NewConfig creates a config from a config path (cPath).
func NewConfig(cPath string) (Config, error) {
	_, err := os.Stat(cPath); if os.IsNotExist(err) {
		return Config{}, err
	}

	confDir, confName := path.Split(cPath)
	name := strings.TrimSuffix(confName, filepath.Ext(confName))

	viper.AddConfigPath(confDir)
	viper.SetConfigName(name)
	viper.SetConfigType("toml")

	err = viper.ReadInConfig(); if err != nil {
		return Config{}, err
	}

	var c Config
	err = viper.Unmarshal(&c)
	if err != nil {
		return Config{}, err
	} else {
		return c, nil
	}
}
package ipagent

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
	"testing"
)

var testEmail = "mail@example.com"

var validConf = fmt.Sprintf(`logging = true
[cloudflare]
zone_id = "aaaaaaaaaaaaaaaa"
api_key = "bbbbbbbbbbbbbbbb"
api_email = "%s"

[domain]
name = "example.com"
proxy = true
type = "A"

[[sub_domains]]
name = "a"
proxy = false
type = "A"

[[sub_domains]]
name = "b"
proxy = true
type = "A"
`, testEmail)

var invalidConf = `lorem ipsum`


func TestUnmarshal(t *testing.T) {
	c, err := configFromString(validConf)
	if err != nil {
		t.Fatalf("Couldn't load config: %v", err)
	}

	if c.Cloudflare.APIEmail != testEmail {
		t.Errorf("config not loaded correctly - email did not match: got '%s', expected: '%s'",
			c.Cloudflare.APIEmail, testEmail)
	}

	if len(c.SubDomains) != 2 {
		t.Errorf("config not loaded correctly - wrong number of subdomains")
	}
}

func TestInvalidLoad(t *testing.T) {
	viper.SetConfigType("toml")
	err := viper.ReadConfig(strings.NewReader(invalidConf))
	if err == nil {
		t.Errorf("did not get error while loading invalid config")
	}
}

func TestConfig_GetDomainList(t *testing.T) {
	c, err := configFromString(validConf)
	if err != nil {
		t.Fatalf("Couldn't load config: %v", err)
	}

	dList := c.GetDomainList()
	if len(dList) != 3 {
		t.Fatalf("Wrong number of domains returned: got %d, expected %d", len(dList), 3)
	}
}

func configFromString(s string) (Config, error) {
	var c Config
	viper.SetConfigType("toml")
	err := viper.ReadConfig(strings.NewReader(s))
	if err != nil {
		return Config{}, nil
	}

	err = viper.Unmarshal(&c)
	if err != nil {
		return Config{}, err
	}

	return c, nil
}
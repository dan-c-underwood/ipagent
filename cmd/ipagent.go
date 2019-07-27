package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"io/ioutil"
	"ipagent"
	"log"
	"log/syslog"
	"net"
	"os"
	"runtime"
)

func main() {
	confPath := flag.String("config", "./ipagent.toml", "Sets the location of the config file")
	dry := flag.Bool("dry", false, "If enabled, application will log but won't update DNS records")

	flag.Parse()

	c, err := ipagent.NewConfig(*confPath); if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Invalid config path specified.")
		} else {
			fmt.Printf("Error loading config: %v", err)
		}
		os.Exit(1)
	}

	if c.Logging && runtime.GOOS != "darwin" {
		logger, err := syslog.New(syslog.LOG_INFO, "ipagent"); if err != nil {
			fmt.Println("Couldn't create syslog handler")
		}
		log.SetOutput(logger)
	}

	mh := NewMsgHandler(c.Logging)

	ip, err := ipagent.QueryIP()
	if err != nil {
		mh.Printf("Error querying IP: %v", err)
		os.Exit(1)
	}
	mh.Printf("Public IP: %s", ip.String())

	cacheIP, err := ReadCache(); if err != nil {
		if !os.IsNotExist(err) {
			mh.Printf("Unable to read cache file due to unexpected error: %v", err)
			os.Exit(1)
		}
	}

	if net.IP.Equal(cacheIP, ip) {
		mh.Printf("Public IP address is same as before, no change required.")
		os.Exit(0)
	}

	err = WriteCache(ip); if err != nil {
		mh.Printf("Unable to write cache file due to error, however execution will continue: %v", err)
	}

	client, err := NewCloudflareClient(c.Cloudflare)
	if err != nil {
		mh.Printf("Unable to create Cloudflare client: %v", err)
		os.Exit(1)
	}

	dMap := make(map[[sha256.Size]byte]cloudflare.DNSRecord)

	records, err := client.GetRecords()
	if err != nil {
		mh.Printf("Unable to retrieve DNS Records: %v", err)
		os.Exit(1)
	}

	mh.Println("Current domains on Cloudflare:")
	for _, record := range records {
		mh.Printf("%s %s: %s, Proxied: %t", record.Type, record.Name, record.Content, record.Proxied)
		dMap[sha256.Sum256([]byte(record.Name + "|" + record.Type))] = record
	}

	mh.Println("Checking for Updates:")
	for _, domain := range c.GetDomainList() {
		record, ok := dMap[sha256.Sum256([]byte(domain.Name + "|" + string(domain.Type)))]; if ok {
			// Update record
			if record.Content != ip.String() {
				mh.Printf("Updating domain '%s' as IP is not current", domain.Name)
				if !*dry {
					err := client.UpdateRecord(domain, record.ID, ip)
					if err != nil {
						mh.Printf("Unable to perform DNS Record update: %v", err)
						os.Exit(1)
					}
				}
			} else {
				mh.Printf("Domain '%s' does not need updating as the IP address is still current", domain.Name)
			}
		} else {
			mh.Printf("Domain '%s' does not currently have record, will create", domain.Name)
			if !*dry {
				err := client.CreateRecord(domain, ip)
				if err != nil {
					mh.Printf("Unable to perform DNS Record create: %v", err)
					os.Exit(1)
				}
			}
		}
	}
}

// CloudflareClient provides a configured wrapper around the Cloudflare API that prevents the need to repeatedly supply
// the same variables.
type CloudflareClient struct {
	api *cloudflare.API
	zoneID string
}

// NewCloudflareClient constructs a new Cloudflare client based on the config.
func NewCloudflareClient(config ipagent.CloudflareConfig) (CloudflareClient, error) {
	api, err := cloudflare.New(config.APIKey, config.APIEmail)
	if err != nil {
		return CloudflareClient{}, err
	}

	return CloudflareClient{api, config.ZoneID}, err
}

// UpdateRecord updates a DNS record in Cloudflare.
func (c CloudflareClient) UpdateRecord(domain ipagent.Domain, recordID string, ip net.IP) error {
	rr := cloudflare.DNSRecord{
		Type:       string(domain.Type),
		Name:       domain.Name,
		Content:	ip.String(),
		Proxied:    domain.Proxy,
	}

	return c.api.UpdateDNSRecord(c.zoneID, recordID, rr)
}

// CreateRecord creates a new DNS record in Cloudflare.
func (c CloudflareClient) CreateRecord(domain ipagent.Domain, ip net.IP) error {
	rr := cloudflare.DNSRecord{
		Type:       string(domain.Type),
		Name:       domain.Name,
		Content:	ip.String(),
		Proxied:    domain.Proxy,
	}

	_, err := c.api.CreateDNSRecord(c.zoneID, rr)
	return err
}

// GetRecords returns all records for the configured zone.
func (c CloudflareClient) GetRecords() ([]cloudflare.DNSRecord, error) {
	rr := cloudflare.DNSRecord{}
	records, err := c.api.DNSRecords(c.zoneID, rr)
	if err != nil {
		return nil, err
	} else {
		return records, nil
	}
}

// MsgHandler is used to output either using Stdout, or through the log package.
type MsgHandler struct {
	logging bool
}

// NewMsgHandler constructs a new MsgHandler, the value of logging dictates whether messages will be output to Stdout
// or via the log package.
func NewMsgHandler(logging bool) MsgHandler {
	return MsgHandler{logging:logging}
}

// Println wraps either the fmt and log Println functions dependent on the configuration.
func (h MsgHandler) Println(msg string) {
	if h.logging {
		log.Println(msg)
	} else {
		fmt.Println(msg)
	}
}

// Printf allows outputting a formatted string via either Stdout or log. It is equivalent to Println(Sprintf())
func (h MsgHandler) Printf(msg string, a ...interface{}) {
	h.Println(fmt.Sprintf(msg, a...))
}

// WriteCache writes an IP address to a cache file in the OS tmp directory. It creates the file if one does not already
// exist.
func WriteCache(ip net.IP) error {
	f, err := os.OpenFile(os.TempDir() + "/ipagent.tmp", os.O_CREATE|os.O_WRONLY, 0664); if err != nil {
		return err
	}

	_, err = f.WriteString(ip.String()); if err != nil {
		return err
	}

	return nil
}

// ReadCache reads from the cache file in the OS tmp directory if it exists. If it returns an error it will be a
// *PathError.
func ReadCache() (net.IP, error) {
	ip, err := ioutil.ReadFile(os.TempDir() + "/ipagent.tmp"); if err != nil {
		return nil, err
	}
	ip = net.ParseIP(string(ip))

	return ip, nil
}
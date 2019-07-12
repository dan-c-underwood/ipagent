package ipagent

import (
	"encoding/json"
	"net"
	"net/http"
)

// IpifyResp is used to unmarshal the JSON response from api.ipify.org.
type IpifyResp struct {
	IP string `json:"ip"`
}

// QueryIP returns a net.IP value that is the current public IP of the network.
func QueryIP() (net.IP, error) {
	res, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		return net.IP{}, err
	}

	ipJson := IpifyResp{}
	err = json.NewDecoder(res.Body).Decode(&ipJson)
	if err != nil {
		return net.IP{}, err
	}

	return net.ParseIP(ipJson.IP), nil
}

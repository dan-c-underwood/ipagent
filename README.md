# ipagent

This project allows for easy updates of the IP address (and proxy settings) associated with DNS records with Cloudflare (provided that IP address is the public IP where this software is running). As a side benefit, it will also create any missing configured DNS records.

## Build

This product should be buildable simply through the use of `go build` for Go >= 1.12.

## Usage

Configuration of this tool is done through the provision of a file called `ipagent.toml` located either in the directory where the binary is located, or specified using the `-config` flag.

A template configuration file is provided for convenience showing the available fields. Additionally, if provided the `-dry` flag will output any actions that would have been taken, however will not update or create records in Cloudflare.

## Future Development

Additional functionality may be added to the project in the future, this includes (a non-exhaustive list):
- [ ] Additional DNS providers (such as DNSSimple)
- [ ] Support for other types of DNS Records (such as TXT)
- [ ] Support for other mechanisms for obtaining the public IP (e.g. from a network interface)

package dns

import (
	"context"
	"net"
)

type DNS struct {
	resolver *net.Resolver
}

func NewDNS(dnsServer string) *DNS {
	return &DNS{
		resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				dialer := net.Dialer{}
				return dialer.DialContext(ctx, "udp", dnsServer)
			},
		},
	}
}

// resolves the given domain to an IP address (DNS lookup)
func (dns *DNS) ResolveIP(domain string) (string, error) {
	ips, err := dns.resolver.LookupHost(context.Background(), domain)
	if err != nil {
		return "", err
	}

	// Return the first IP address
	return ips[0], nil
}

// resolves the given IP address to a domain name (reverse DNS lookup)
func (dns *DNS) ResolveDomain(ip string) (string, error) {
	// Use the custom resolver to resolve the IP address to a domain
	domains, err := dns.resolver.LookupAddr(context.Background(), ip)
	if err != nil {
		return "", err
	}

	return domains[0], nil
}

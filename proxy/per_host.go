// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//go:generate safemap -k string -v Dialer -n proxy

package proxy

import (
	"net"
	"strings"
)

// A PerHost directs connections to a default Dialer unless the hostname
// requested matches one of a number of exceptions.
type PerHost struct {
	def, bypass Dialer

	cache *proxySafeMap

	bypassCIDRs    []*net.IPNet
	bypassIPs      []net.IP
	bypassKEYWORDs []string
	bypassDOMAINs  []string
	bypassSUFFIXs  []string
}

// NewPerHost returns a PerHost Dialer that directs connections to either
// defaultDialer or bypass, depending on whether the connection matches one of
// the configured rules.
func NewPerHost(defaultDialer, bypass Dialer) *PerHost {
	cache := NewproxySafeMap(nil)
	return &PerHost{
		def:    defaultDialer,
		bypass: bypass,
		cache:  cache,
	}
}

// Dial connects to the address addr on the given network through either
// defaultDialer or bypass.
func (p *PerHost) Dial(network, addr string) (c net.Conn, err error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	return p.dialerForRequest(host).Dial(network, addr)
}

// getDialerByRule return the Dialer to use
func (p *PerHost) getDialerByRule(host string) Dialer {
	for _, domains := range p.bypassDOMAINs {
		if domains == host {
			return p.bypass
		}
	}

	for _, suffix := range p.bypassSUFFIXs {
		if strings.HasSuffix(host, suffix) {
			return p.bypass
		}
	}

	for _, keyword := range p.bypassKEYWORDs {
		if strings.Contains(host, keyword) {
			return p.bypass
		}
	}

	if ip := net.ParseIP(host); ip != nil {
		for _, bypassIP := range p.bypassIPs {
			if bypassIP.Equal(ip) {
				return p.bypass
			}
		}

		for _, net := range p.bypassCIDRs {
			if net.Contains(ip) {
				return p.bypass
			}
		}
	}
	return p.def
}

// a cache wrapper to getDialerByRule
func (p *PerHost) dialerForRequest(host string) Dialer {
	d, ok := p.cache.Get(host)

	if !ok {
		dialer := p.getDialerByRule(host)
		p.cache.Set(host, dialer)
		return dialer
	}

	return d

}

// AddFromString parses a string that contains comma-separated values
// specifying hosts that should use the bypass proxy. Each value is either an
// IP address, a CIDR range, a zone (*.example.com) or a hostname
// (localhost). A best effort is made to parse the string and errors are
// ignored.
func (p *PerHost) AddFromString(s string) {
	hosts := strings.Split(s, ",")
	switch hosts[0] {
	case "DOMAIN":
		p.AddDOMAIN(hosts[1])
	case "IP":
		if ip := net.ParseIP(hosts[1]); ip != nil {
			p.AddIP(ip)
		}
	case "IP-CIDR":
		if _, net, err := net.ParseCIDR(hosts[1]); err == nil {
			p.AddCIDR(net)
		}
	case "DOMAIN-SUFFIX":
		p.AddSUFFIX(hosts[1])
	case "DOMAIN-KEYWORD":
		p.AddKEYWORD(hosts[1])
	}
}

// AddIP specifies an IP address that will use the bypass proxy. Note that
// this will only take effect if a literal IP address is dialed. A connection
// to a named host will never match an IP.
func (p *PerHost) AddIP(ip net.IP) {
	p.bypassIPs = append(p.bypassIPs, ip)
}

// AddCIDR specifies an IP address that will use the bypass proxy. Note that
// this will only take effect if a literal IP address is dialed. A connection
// to a named host will never match an IP.
func (p *PerHost) AddCIDR(net *net.IPNet) {
	p.bypassCIDRs = append(p.bypassCIDRs, net)
}

// AddKEYWORD specifies an IP address that will use the bypass proxy. Note that
// this will only take effect if a literal IP address is dialed. A connection
// to a named host will never match an IP.
func (p *PerHost) AddKEYWORD(keyword string) {
	p.bypassKEYWORDs = append(p.bypassKEYWORDs, keyword)
}

// AddDOMAIN specifies an IP address that will use the bypass proxy. Note that
// this will only take effect if a literal IP address is dialed. A connection
// to a named host will never match an IP.
func (p *PerHost) AddDOMAIN(domain string) {
	p.bypassDOMAINs = append(p.bypassDOMAINs, domain)
}

// AddSUFFIX specifies an IP address that will use the bypass proxy. Note that
// this will only take effect if a literal IP address is dialed. A connection
// to a named host will never match an IP.
func (p *PerHost) AddSUFFIX(suffix string) {
	p.bypassSUFFIXs = append(p.bypassSUFFIXs, suffix)
}

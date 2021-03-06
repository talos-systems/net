// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package net_test

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	talosnet "github.com/talos-systems/net"
)

func TestEmpty(t *testing.T) {
	// added for accurate coverage estimation
	//
	// please remove it once any unit-test is added
	// for this package
}

func TestAddressContainsPort(t *testing.T) {
	assert.Equal(t, talosnet.AddressContainsPort("192.168.1.1:9021"), true)
	assert.Equal(t, talosnet.AddressContainsPort("node0:10001"), true)
	assert.Equal(t, talosnet.AddressContainsPort("node0.testdomain.io:10001"), true)
	assert.Equal(t, talosnet.AddressContainsPort("[2001:db8::1]:64321"), true)
	assert.Equal(t, talosnet.AddressContainsPort("[2001:db8:3:4:5:6:7:1]:64321"), true)

	assert.Equal(t, talosnet.AddressContainsPort("[2001:db8:3:4:5:6:7:1:bad]:64321"), false)
	assert.Equal(t, talosnet.AddressContainsPort("2001:db8:0::2000"), false)
	assert.Equal(t, talosnet.AddressContainsPort("::1"), false)
	assert.Equal(t, talosnet.AddressContainsPort("127.0.0.1"), false)
	assert.Equal(t, talosnet.AddressContainsPort("node0"), false)
	assert.Equal(t, talosnet.AddressContainsPort("node0.testdomain.io"), false)
}

func TestFormatAddress(t *testing.T) {
	assert.Equal(t, talosnet.FormatAddress("2001:db8::1"), "[2001:db8::1]")
	assert.Equal(t, talosnet.FormatAddress("[2001:db8::1]"), "[2001:db8::1]")
	assert.Equal(t, talosnet.FormatAddress("192.168.1.1"), "192.168.1.1")
	assert.Equal(t, talosnet.FormatAddress("alpha.beta.gamma.com"), "alpha.beta.gamma.com")
}

func TestFormatCIDR(t *testing.T) {
	ip4 := net.ParseIP("192.168.1.1")
	_, cidr4, _ := net.ParseCIDR("192.168.0.0/16") //nolint: errcheck

	ip6 := net.ParseIP("2001:db8::1")
	_, cidr6, _ := net.ParseCIDR("2001:db8::/32") //nolint: errcheck

	assert.Equal(t, talosnet.FormatCIDR(ip4, *cidr4), "192.168.1.1/16")
	assert.Equal(t, talosnet.FormatCIDR(ip6, *cidr6), "2001:db8::1/32")
}

//nolint: scopelint
func TestNthIPInNetwork(t *testing.T) {
	type args struct {
		network *net.IPNet
		n       int
	}

	tests := []struct {
		name string
		args args
		want net.IP
	}{
		{
			name: "increment IPv4 by 1",
			args: args{
				network: &net.IPNet{
					IP:   net.IP{10, 96, 0, 0},
					Mask: net.IPMask{255, 255, 255, 0},
				},
				n: 1,
			},
			want: net.IP{10, 96, 0, 1},
		},
		{
			name: "increment IPv4 by 10",
			args: args{
				network: &net.IPNet{
					IP:   net.IP{10, 96, 0, 0},
					Mask: net.IPMask{255, 255, 255, 0},
				},
				n: 10,
			},
			want: net.IP{10, 96, 0, 10},
		},
		{
			name: "increment IPv6 by 1",
			args: args{
				network: &net.IPNet{
					IP:   net.ParseIP("2001:db8:a0b:12f0::1"),
					Mask: net.IPMask{255, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				n: 1,
			},
			want: net.ParseIP("2001:db8:a0b:12f0::2"),
		},
		{
			name: "increment IPv6 by 10",
			args: args{
				network: &net.IPNet{
					IP:   net.ParseIP("2001:db8:a0b:12f0::1"),
					Mask: net.IPMask{255, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				n: 10,
			},
			want: net.ParseIP("2001:db8:a0b:12f0::b"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := talosnet.NthIPInNetwork(tt.args.network, tt.args.n)
			if err != nil {
				t.Errorf("%w", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NthFromIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateEndpointURI(t *testing.T) {
	goodTests := []string{
		"http://216.22.102.222",
		"https://[2001:db8:abef::ffff]:65000",
		"https://[2001:db8::ffff]",
		"http://goethe.de",
	}

	for _, testEP := range goodTests {
		assert.Nil(t, talosnet.ValidateEndpointURI(testEP), "URI should be valid")
	}

	badTests := []string{
		"12.34.56.89:1234",                 // ipv4:port, no protocol
		"[2001:db8::1]:5040",               // ipv6:port, no protocol
		"hostA:65301",                      // host:port, no protocol
		"my.long.domain.name:10101",        // dns:port, no protocol
		"192.168.2.1",                      // IP without port
		"[2001:db8::1]",                    // IPv6 without port
		"kubernetes.io",                    // hostname without port
		"2001:db8:123:445:204",             // IPv6 without brackets
		"http://2001:db8:101:101::1:50000", // IPv6 URL without brackets
		"http://192.168.1.1:1020304",       // Port out of range
		"http://192.168.1.1:0",             // 0 Port
		"http://192.168.1.1:-1000",         // Negative Port
	}

	for _, testEP := range badTests {
		assert.NotNil(t, talosnet.ValidateEndpointURI(testEP), "URI should be invalid")
	}
}

func TestParseCIDR(t *testing.T) {
	goodTests := map[string]*net.IPNet{
		"10.66.0.66": {
			IP:   net.ParseIP("10.66.0.66"),
			Mask: net.CIDRMask(32, 32),
		},
		"10.66.0.66/24": {
			IP:   net.ParseIP("10.66.0.66"),
			Mask: net.CIDRMask(24, 32),
		},
		"2001:db8:abef::ffff": {
			IP:   net.ParseIP("2001:db8:abef::ffff"),
			Mask: net.CIDRMask(128, 128),
		},
		"2001:db8:abef::ffff/32": {
			IP:   net.ParseIP("2001:db8:abef::ffff"),
			Mask: net.CIDRMask(32, 128),
		},
		"[2001:db8:abef::ffff]/32": {
			IP:   net.ParseIP("2001:db8:abef::ffff"),
			Mask: net.CIDRMask(32, 128),
		},
	}

	for in, expected := range goodTests {
		parsedIP, err := talosnet.ParseCIDR(in)
		assert.Nil(t, err, "error should be nil")
		assert.True(t, parsedIP.IP.Equal(expected.IP), "IP addresses should be equal")
		assert.Equal(t, expected.Mask.String(), parsedIP.Mask.String(), "Network masks should be equal")
	}

	badTests := []string{
		"hostname.domain.org",        // name instead of IP
		"http://hostname.domain.org", // URL instead of IP
		"12.34.56.89:1234",           //  IP + port
		"12.34.56.89/96",             //  Subnet mask out of range for family
		"12.34.56.89/96",             //  Subnet mask out of range for family
		"[2001:db8::1]:5040",         // ipv6 + port
	}

	for _, in := range badTests {
		_, err := talosnet.ParseCIDR(in)
		assert.NotNil(t, err, fmt.Sprintf("ParseCIDR(%s) should return an error", in))
	}
}

package webkit

import "testing"

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		// 10.0.0.0/8 range
		{name: "10.0.0.1 is private", ip: "10.0.0.1", want: true},
		{name: "10.255.255.255 is private", ip: "10.255.255.255", want: true},

		// 172.16.0.0/12 range
		{name: "172.16.0.1 is private", ip: "172.16.0.1", want: true},
		{name: "172.31.255.255 is private", ip: "172.31.255.255", want: true},
		{name: "172.15.0.1 is not private", ip: "172.15.0.1", want: false},
		{name: "172.32.0.1 is not private", ip: "172.32.0.1", want: false},

		// 192.168.0.0/16 range
		{name: "192.168.0.1 is private", ip: "192.168.0.1", want: true},
		{name: "192.168.255.255 is private", ip: "192.168.255.255", want: true},

		// Public IPs
		{name: "8.8.8.8 is not private", ip: "8.8.8.8", want: false},
		{name: "1.1.1.1 is not private", ip: "1.1.1.1", want: false},

		// Edge cases
		{name: "empty string is not private", ip: "", want: false},
		{name: "invalid IP is not private", ip: "not-an-ip", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isPrivateIP(tc.ip)
			if got != tc.want {
				t.Errorf("isPrivateIP(%q) = %v, want %v", tc.ip, got, tc.want)
			}
		})
	}
}

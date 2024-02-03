package ssh_client

import (
	"path/filepath"
	"testing"
)

func Test_parseSshURI(t *testing.T) {
	testData := []struct {
		URI  string
		Host string
		Port string
		User string
	}{
		{
			URI:  "u@h:p",
			Host: "h",
			Port: "p",
			User: "u",
		},
		{
			URI:  "ssh://u@h:p",
			Host: "h",
			Port: "p",
			User: "u",
		},
		{
			URI:  "h:p",
			Host: "h",
			Port: "p",
		},
		{
			URI:  "u@h",
			Host: "h",
			User: "u",
		},
		{
			URI:  "h",
			Host: "h",
		},
	}

	for _, tc := range testData {
		t.Run(tc.URI, func(t *testing.T) {
			u, h, p := parseSshURI(tc.URI)
			if u != tc.User ||
				h != tc.Host ||
				p != tc.Port {
				t.Errorf("Failed to parse URI into (user,host,port) got (%s,%s,%s) want (%s,%s,%s)", u, h, p, tc.User, tc.Host, tc.Port)
			}
		})
	}
}

func Test_configure(t *testing.T) {
	testData := []struct {
		URI  string
		Host string
		Port string
		User string
	}{
		{
			URI:  "u@h:p",
			Host: "h",
			Port: "p",
			User: "u",
		},
		{
			URI:  "ssh://u@h:p",
			Host: "h",
			Port: "p",
			User: "u",
		},
		{
			URI:  "h:p",
			Host: "h",
			Port: "p",
		},
		{
			URI:  "u@h",
			Host: "h",
			User: "u",
			Port: "22",
		},
		{
			URI:  "h",
			Host: "h",
			Port: "22",
		},
		{
			URI:  "configured-myhost",
			Host: "myhost",
			Port: "2222",
			User: "remoteuser",
		},
	}

	cfgfile, err := filepath.Abs("testdata/ssh_config")
	if err != nil {
		t.Fatalf("cannot read testdata/ssh_config: %v", err)
	}
	cfg, err := OpenSSHConfig(cfgfile)
	if err != nil {
		t.Fatalf("cannot read testdata/ssh_config: %v", err)
	}

	for _, tc := range testData {
		t.Run(tc.URI, func(t *testing.T) {
			h := ParseSshURI(tc.URI)
			var called bool
			h.configure(cfg, func() string {
				called = true
				return ""
			})
			if called && tc.User != "" {
				t.Error("currentUsername called while it should not")
			}
			if !called && tc.User == "" {
				t.Error("currentUsername not called while it should have")
			}
			if h.User != tc.User ||
				h.Hostname != tc.Host ||
				h.Port != tc.Port {
				t.Errorf("Failed to parse URI into (user,host,port) got (%s,%s,%s) want (%s,%s,%s)", h.User, h.Hostname, h.Port, tc.User, tc.Host, tc.Port)
			}
		})
	}

}

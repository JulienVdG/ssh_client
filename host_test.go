package ssh_client

import "testing"

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

package ssh_client

import "testing"

func TestExpandTokens(t *testing.T) {
	testData := []struct {
		In  string
		Out string
	}{
		{
			In:  "bla",
			Out: "bla",
		},
		{
			In:  "bla%",
			Out: "bla%",
		},
		{
			In:  "%n to %r@%h:%p",
			Out: "Name to remoteUser@Host:2222",
		},
		{
			In:  "bla%%d",
			Out: "bla%d",
		},
		{
			In:  "Invalid %# Token",
			Out: "Invalid %# Token",
		},
	}

	h := &Host{
		Name:     "Name",
		Hostname: "Host",
		Port:     "2222",
		User:     "remoteUser",
	}

	for _, tc := range testData {
		t.Run(tc.In, func(t *testing.T) {
			out := h.ExpandTokens(tc.In)
			if out != tc.Out {
				t.Errorf("Failed to apply token got \"%s\", want \"%s\"", out, tc.Out)
			}
		})
	}
}

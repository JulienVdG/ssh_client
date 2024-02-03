package ssh_client

import "testing"

func Test_expandHome(t *testing.T) {
	testData := []struct {
		In   string
		Home string
		User string
		Out  string
	}{
		{
			In:  "toto/titi",
			Out: "toto/titi",
		},
		{
			In:   "~u/toto/titi",
			Home: "/home",
			User: "u",
			Out:  "/home/toto/titi",
		},
		{
			In:   "~/toto/titi",
			Home: "/home",
			Out:  "/home/toto/titi",
		},
		{
			In:   "~",
			Home: "/home",
			Out:  "/home",
		},
		{
			In:   "~/",
			Home: "/home",
			Out:  "/home",
		},
		{
			In:   "~u",
			Home: "/home",
			User: "u",
			Out:  "/home",
		},
		{
			In:   "~u/",
			Home: "/home",
			User: "u",
			Out:  "/home",
		},
		{
			In:   "~u/toto",
			User: "u",
			Out:  "~u/toto",
		},
	}

	for _, tc := range testData {
		t.Run(tc.In, func(t *testing.T) {
			var homedircalled, userhomedircalled bool
			out := expandHome(tc.In,
				func() string {
					homedircalled = true
					return tc.Home
				},
				func(user string) string {
					userhomedircalled = true
					if user != tc.User {
						t.Errorf("Failed to match user got %s want %s", user, tc.User)
					}
					return tc.Home
				},
			)
			if homedircalled && (tc.User != "" || tc.Home == "") {
				t.Error("homedir called while it should not")
			}
			if userhomedircalled && tc.User == "" && tc.Home == "" {
				t.Error("userhomedir called while it should not")
			}
			if out != tc.Out {
				t.Errorf("Failed to expand tilde got %s want %s", out, tc.Out)
			}
		})
	}
}

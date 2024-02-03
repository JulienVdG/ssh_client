package ssh_client

import (
	"os"
	osuser "os/user"
	"path/filepath"
	"strings"
)

func homedir() string {
	user, err := osuser.Current()
	if err == nil {
		return user.HomeDir
	} else {
		return os.Getenv("HOME")
	}
}

func currentUsername() string {
	user, err := osuser.Current()
	if err == nil {
		return user.Username
	} else {
		return os.Getenv("USER")
	}
}

func userdir(user string) string {
	u, err := osuser.Lookup(user)
	if err != nil {
		return ""
	}
	return u.HomeDir
}

func expandHome(inputPath string, homedir func() string, userhomedir func(string) string) string {
	if len(inputPath) == 0 {
		return inputPath
	}
	if inputPath[0] != '~' {
		return inputPath
	}
	path := filepath.ToSlash(inputPath[1:])
	var user string
	slashIdx := strings.IndexByte(path, '/')
	if slashIdx == -1 {
		user = path
		path = ""
	} else {
		user = path[:slashIdx]
		path = path[slashIdx+1:]
	}

	var home string
	if user == "" {
		home = homedir()
	} else {
		home = userhomedir(user)
		if home == "" {
			return inputPath
		}
	}

	return filepath.Join(home, filepath.FromSlash(path))
}

func ExpandHome(inputPath string) string {
	return expandHome(inputPath, homedir, userdir)
}

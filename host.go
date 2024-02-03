package ssh_client

import (
	"strings"

	"github.com/kevinburke/ssh_config"
)

type Host struct {
	Name     string
	Hostname string
	Port     string
	User     string
}

func parseSshURI(uri string) (user, host, port string) {
	// TODO: do we want regex here to validate the exact format ?
	userHostPort := strings.TrimPrefix(uri, "ssh://")
	atIdx := strings.IndexByte(userHostPort, '@')
	if atIdx != -1 {
		user = userHostPort[:atIdx]
	}
	hostPort := userHostPort[atIdx+1:]
	colonIdx := strings.IndexByte(hostPort, ':')
	if colonIdx != -1 {
		port = hostPort[colonIdx+1:]
		host = hostPort[:colonIdx]
	} else {
		host = hostPort
	}

	return
}

// ParseSshURI parse either [user@]host[:port] or an ssh URI ie ssh://[user@]host[:port].
func ParseSshURI(uri string) *Host {
	u, h, p := parseSshURI(uri)
	return &Host{
		Name:     h,
		Hostname: h,
		Port:     p,
		User:     u,
	}
}

func (h *Host) Configure(u *ssh_config.UserSettings) error {
	if h.Port == "" {
		h.Port = u.Get(h.Name, "Port")
	}
	h.Hostname = u.Get(h.Name, "Hostname")
	// Default to Name
	if h.Hostname == "" {
		h.Hostname = h.Name
	}
	if h.User == "" {
		h.User = u.Get(h.Name, "User")
	}
	return nil
}

func (h *Host) Addr() string {
	return h.Hostname + ":" + h.Port
}

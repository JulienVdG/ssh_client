package ssh_client

import (
	"strings"
)

type sshSettingsGetter interface {
	Get(alias, key string) string
	GetAll(alias, key string) []string
}

type Host struct {
	Name     string
	Hostname string
	Port     string
	User     string
	cfg      sshSettingsGetter
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

func (h *Host) configure(cfg sshSettingsGetter, currentUsername func() string) error {
	h.cfg = cfg
	if h.Port == "" {
		h.Port = cfg.Get(h.Name, "Port")
	}
	h.Hostname = cfg.Get(h.Name, "Hostname")
	// Default to Name
	if h.Hostname == "" {
		h.Hostname = h.Name
	}
	if h.User == "" {
		h.User = cfg.Get(h.Name, "User")
	}
	if h.User == "" {
		h.User = currentUsername()
	}
	return nil
}

func (h *Host) Configure(cfg sshSettingsGetter) error {
	return h.configure(cfg, currentUsername)
}

func (h *Host) Addr() string {
	return h.Hostname + ":" + h.Port
}

func (h *Host) ConfigGet(key string) string {
	return h.cfg.Get(h.Name, key)
}

func (h *Host) ConfigGetAll(key string) []string {
	return h.cfg.GetAll(h.Name, key)
}

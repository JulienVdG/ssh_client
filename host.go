package ssh_client

import (
	"errors"
	"fmt"

	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
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

func (h *Host) KnownHosts() []string {
	u := h.ConfigGet("UserKnownHostsFile")
	f := strings.Split(u, " ")
	for i := range f {
		f[i] = h.ExpandTokens(ExpandHome(f[i]))
	}
	g := h.ConfigGet("GlobalKnownHostsFile")
	f = append(f, strings.Split(g, " ")...)
	var knownHostFiles []string
	for i := range f {
		info, err := os.Stat(f[i])
		if err != nil {
			continue
		}
		if info.Mode().IsRegular() {
			knownHostFiles = append(knownHostFiles, f[i])
		}
	}
	return knownHostFiles
}

func (h *Host) AgentSockName() string {
	n := h.ConfigGet("IdentityAgent")
	// ssh_config does not handle IdentityAgent and its SSH_AUTH_SOCK default.
	if n == "" {
		n = "SSH_AUTH_SOCK"
	}

	n = ExpandHome(n)
	n = h.ExpandTokens(n)
	n = os.ExpandEnv(n)
	if n == "SSH_AUTH_SOCK" {
		n = os.Getenv("SSH_AUTH_SOCK")
	}
	return n
}

var ErrAgentDisabled = errors.New("SSH Agent Disabled")

// TODO once or cache
func (h *Host) Agent() (agent.ExtendedAgent, error) {
	socket := h.AgentSockName()
	if strings.ToLower(socket) == "none" {
		return nil, ErrAgentDisabled
	}
	agentConn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, fmt.Errorf("Failed to open Agent (on %s): %w", socket, err)
	}
	agentClient := agent.NewClient(agentConn)
	return agentClient, nil
}

func (h *Host) IdentitiesOnly() bool {
	switch strings.ToLower(h.ConfigGet("IdentitiesOnly")) {
	case "true", "yes":
		return true
	case "false", "no":
		return false
	default:
		// TODO: fail or at least log
		return false
	}
}

func (h *Host) IdentityPublicKeys() []ssh.PublicKey {
	identityFiles := h.ConfigGetAll("IdentityFile")
	var pubKeys []ssh.PublicKey
	for _, file := range identityFiles {
		fname := ExpandHome(file)
		fname = h.ExpandTokens(fname)
		fname += ".pub"
		content, err := os.ReadFile(fname)
		if err != nil {
			// TODO slog and only when not from DefaultValues
			// fmt.Printf("could not read %s: %v\n", fname, err)
			continue
		}

		k, _, _, _, err := ssh.ParseAuthorizedKey(content)
		if err != nil {
			// TODO slog
			fmt.Printf("could not parse %s: %v\n%s\n", fname, err, string(content))
			continue
		}
		pubKeys = append(pubKeys, k)
	}
	return pubKeys
}

func (h *Host) GetSignersCallback() (func() ([]ssh.Signer, error), error) {
	agentClient, err := h.Agent()
	if err != nil {
		return nil, fmt.Errorf("SSH Agent Required: %w", err)
	}
	if !h.IdentitiesOnly() {
		// Ignore identities from config, loading privateKeys is not implemented, assume the required ones are loaded.
		return agentClient.Signers, nil
	}

	// IdentitiesOnly=yes so we need to filter agent Signers according to identities in config.
	hasKeys := make(map[string]bool)
	for _, k := range h.IdentityPublicKeys() {
		hash := string(k.Marshal())
		hasKeys[hash] = true
	}

	cb := func() ([]ssh.Signer, error) {
		agentSigners, err := agentClient.Signers()
		if err != nil {
			return nil, err
		}
		var signers []ssh.Signer
		for i := range agentSigners {
			hash := string(agentSigners[i].PublicKey().Marshal())
			if hasKeys[hash] {
				//fmt.Println("found matching pubkey in agent")
				signers = append(signers, agentSigners[i])
			}
		}
		return signers, nil
	}
	return cb, nil
}
func (h *Host) GetSigners() ([]ssh.Signer, error) {
	cb, err := h.GetSignersCallback()
	if err != nil {
		return nil, err
	}
	return cb()
}

func (h *Host) ClientConfig() (*ssh.ClientConfig, error) {
	hostKeyCallback, err := knownhosts.New(h.KnownHosts()...)
	if err != nil {
		return nil, fmt.Errorf("could not create host key callback: %w", err)
	}
	getSigners, err := h.GetSignersCallback()
	if err != nil {
		return nil, fmt.Errorf("could not create key signers callback: %w", err)
	}
	cfg := &ssh.ClientConfig{
		User: h.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(getSigners),
		},
		HostKeyCallback: hostKeyCallback,
	}

	return cfg, nil
}

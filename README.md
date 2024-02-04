# ssh_client
Go ssh client configured from openssh config files

## Go code example

```
	host, err := ssh_client.New("ProxyChainHost", ssh_client.DefaultUserSettings)
	if err != nil {
		log.Fatalln(err)
	}

	conn, err := host.Dial("tcp")
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	// From here we can use `conn` to Dial remote ports and/or open a session.
```

## Corresponding `~/.ssh/config` example

```
Host ProxyChainHost
	Hostname FinalHostname
	ProxyJump %r@FirstProxy,ConfiguredProxy

Host ConfiguredProxy
	Hostname intnames.dedibox.vdg.name
	User root
```

## Notes
 - The keys required for all hosts must be loaded in ssh-agent _before_ connection. (loading keys is not supported).
 - Entries in `known_hosts` must be present (Adding entries is not supported).

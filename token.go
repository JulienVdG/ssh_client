package ssh_client

import (
	"strings"
)

// ExpandTokens expand % tokens
//
// %%    A literal ‘%’.
// %h    The remote hostname.
// %n    The original remote hostname, as given on the command line.
// %p    The remote port.
// %r    The remote username.
//
// TODO: complete the list of tokens and allow selecting them depending on the key
// see: https://github.com/openssh/openssh-portable/blob/master/sshconnect.h#L54
// %c see https://github.com/openssh/openssh-portable/blob/master/readconf.c#L354
func (h *Host) ExpandTokens(in string) string {
	var out string
	for {
		percentIdx := strings.IndexByte(in, '%')
		if percentIdx == -1 {
			// No %
			break
		}
		if percentIdx+1 >= len(in) {
			// Last char is %
			break
		}
		out = out + in[:percentIdx]
		token := in[percentIdx+1]
		in = in[percentIdx+2:]
		switch token {
		case '%':
			out = out + "%"
		case 'h':
			out = out + h.Hostname
		case 'n':
			out = out + h.Name
		case 'p':
			out = out + h.Port
		case 'r':
			out = out + h.User
		default:
			// TODO: unknown token should err?
			out = out + "%" + string(token)
		}
	}
	out = out + in
	return out
}

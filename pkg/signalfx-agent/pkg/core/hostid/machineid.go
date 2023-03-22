package hostid

import (
	systemdutil "github.com/coreos/go-systemd/util"
)

// MachineID returns the Linux machine-id, which is present on most newer Linux
// distros.  It is more useful than hostname as a unique identifier.  See
// http://man7.org/linux/man-pages/man5/machine-id.5.html
func MachineID() string {
	mid, err := systemdutil.GetMachineID()
	if err != nil {
		return ""
	}
	return mid
}

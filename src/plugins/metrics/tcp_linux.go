package metrics

import (
	"os/exec"
)

const REBOOT_CMD = "/system/bin/reboot"

func (tcp *TcpStats) performReboot() {
	cmd := exec.Command(REBOOT_CMD)
	cmd.Start()
}

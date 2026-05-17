//go:build !windows

package audio

import "syscall"

func execSignal(name string) syscall.Signal {
	if name == "CONT" {
		return syscall.SIGCONT
	}
	return syscall.SIGSTOP
}

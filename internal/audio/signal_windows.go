//go:build windows

package audio

import "os"

func execSignal(string) os.Signal {
	return os.Kill
}

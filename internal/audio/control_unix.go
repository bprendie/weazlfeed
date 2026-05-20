//go:build !windows

package audio

import (
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type socketControl struct {
	path string
}

func newMPVControl() mpvControl {
	name := "weazlfeed-" + strconv.Itoa(os.Getpid()) + "-" + strconv.FormatInt(time.Now().UnixNano(), 10) + ".sock"
	return &socketControl{path: filepath.Join(os.TempDir(), name)}
}

func (c *socketControl) args() []string {
	return []string{"--input-ipc-server=" + c.path}
}

func (c *socketControl) command(command string) error {
	conn, err := net.DialTimeout("unix", c.path, time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write([]byte(command))
	return err
}

func (c *socketControl) close() {
	_ = os.Remove(c.path)
}

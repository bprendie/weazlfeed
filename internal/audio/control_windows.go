//go:build windows

package audio

import (
	"os"
	"strconv"
	"time"
)

type namedPipeControl struct {
	path string
}

func newMPVControl() mpvControl {
	name := "weazlfeed-" + strconv.Itoa(os.Getpid()) + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	return &namedPipeControl{path: `\\.\pipe\` + name}
}

func (c *namedPipeControl) args() []string {
	return []string{"--input-ipc-server=" + c.path}
}

func (c *namedPipeControl) command(command string) error {
	file, err := os.OpenFile(c.path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(command)
	return err
}

func (c *namedPipeControl) close() {}

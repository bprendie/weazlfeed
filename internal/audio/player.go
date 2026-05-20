package audio

import (
	"errors"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

type Player struct {
	cmd        *exec.Cmd
	control    mpvControl
	startedAt  time.Time
	offset     int
	pausedAt   time.Time
	pausedTime time.Duration
	paused     bool
	mu         sync.Mutex
}

func Available() bool {
	_, err := exec.LookPath("mpv")
	return err == nil
}

func (p *Player) Play(url string, offsetSeconds int) error {
	bin, err := exec.LookPath("mpv")
	if err != nil {
		return errors.New("mpv not found")
	}
	p.Stop()
	control := newMPVControl()
	args := []string{"--no-video", "--really-quiet"}
	args = append(args, "--input-terminal=no")
	args = append(args, control.args()...)
	if offsetSeconds > 0 {
		args = append(args, "--start="+strconv.Itoa(offsetSeconds))
	}
	args = append(args, url)
	cmd := exec.Command(bin, args...)
	if err := cmd.Start(); err != nil {
		control.close()
		return err
	}
	p.mu.Lock()
	p.cmd = cmd
	p.control = control
	p.startedAt = time.Now()
	p.offset = offsetSeconds
	p.pausedAt = time.Time{}
	p.pausedTime = 0
	p.paused = false
	p.mu.Unlock()
	go func() {
		_ = cmd.Wait()
		control.close()
		p.mu.Lock()
		if p.cmd == cmd {
			p.cmd = nil
			p.control = nil
		}
		p.mu.Unlock()
	}()
	return nil
}

func (p *Player) Seek(deltaSeconds int) error {
	p.mu.Lock()
	url := ""
	if p.cmd != nil && len(p.cmd.Args) > 0 {
		url = p.cmd.Args[len(p.cmd.Args)-1]
	}
	next := p.positionLocked() + deltaSeconds
	p.mu.Unlock()
	if url == "" {
		return nil
	}
	if next < 0 {
		next = 0
	}
	return p.Play(url, next)
}

func (p *Player) TogglePause() error {
	p.mu.Lock()
	cmd := p.cmd
	control := p.control
	if !p.paused {
		p.pausedAt = time.Now()
		p.paused = true
	}
	p.mu.Unlock()
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	if control == nil {
		return errors.New("mpv control pipe is not available")
	}
	return control.command(`{"command":["cycle","pause"]}` + "\n")
}

func (p *Player) Resume() error {
	p.mu.Lock()
	cmd := p.cmd
	control := p.control
	if p.paused {
		p.pausedTime += time.Since(p.pausedAt)
		p.pausedAt = time.Time{}
		p.paused = false
	}
	p.mu.Unlock()
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	if control == nil {
		return errors.New("mpv control pipe is not available")
	}
	return control.command(`{"command":["cycle","pause"]}` + "\n")
}

func (p *Player) Stop() {
	p.mu.Lock()
	cmd := p.cmd
	control := p.control
	p.cmd = nil
	p.control = nil
	p.mu.Unlock()
	if cmd != nil && cmd.Process != nil {
		if control != nil {
			_ = control.command(`{"command":["quit"]}` + "\n")
		}
		_ = cmd.Process.Kill()
	}
}

func (p *Player) Position() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.positionLocked()
}

func (p *Player) positionLocked() int {
	if p.cmd == nil || p.startedAt.IsZero() {
		return p.offset
	}
	elapsed := time.Since(p.startedAt) - p.pausedTime
	if p.paused && !p.pausedAt.IsZero() {
		elapsed -= time.Since(p.pausedAt)
	}
	return p.offset + int(elapsed.Seconds())
}

func (p *Player) Active() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cmd != nil
}

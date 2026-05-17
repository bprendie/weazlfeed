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
	args := []string{"--no-video", "--really-quiet"}
	if offsetSeconds > 0 {
		args = append(args, "--start="+strconv.Itoa(offsetSeconds))
	}
	args = append(args, url)
	cmd := exec.Command(bin, args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	p.mu.Lock()
	p.cmd = cmd
	p.startedAt = time.Now()
	p.offset = offsetSeconds
	p.pausedAt = time.Time{}
	p.pausedTime = 0
	p.paused = false
	p.mu.Unlock()
	go func() {
		_ = cmd.Wait()
		p.mu.Lock()
		if p.cmd == cmd {
			p.cmd = nil
		}
		p.mu.Unlock()
	}()
	return nil
}

func (p *Player) TogglePause() error {
	p.mu.Lock()
	cmd := p.cmd
	if !p.paused {
		p.pausedAt = time.Now()
		p.paused = true
	}
	p.mu.Unlock()
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	return cmd.Process.Signal(execSignal("STOP"))
}

func (p *Player) Resume() error {
	p.mu.Lock()
	cmd := p.cmd
	if p.paused {
		p.pausedTime += time.Since(p.pausedAt)
		p.pausedAt = time.Time{}
		p.paused = false
	}
	p.mu.Unlock()
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	return cmd.Process.Signal(execSignal("CONT"))
}

func (p *Player) Stop() {
	p.mu.Lock()
	cmd := p.cmd
	p.cmd = nil
	p.mu.Unlock()
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}

func (p *Player) Position() int {
	p.mu.Lock()
	defer p.mu.Unlock()
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

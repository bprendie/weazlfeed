package audio

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os/exec"
	"strconv"
	"sync"
)

type Sample struct {
	Level     float64
	Transient float64
	Bands     []float64
	Live      bool
}

type Meter struct {
	cmd  *exec.Cmd
	done chan struct{}
	out  chan Sample
	mu   sync.Mutex
}

func StartMeter(url string, offsetSeconds int) (*Meter, error) {
	bin, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, errors.New("ffmpeg not found")
	}
	args := []string{"-nostdin", "-v", "error"}
	if offsetSeconds > 0 {
		args = append(args, "-ss", strconv.Itoa(offsetSeconds))
	}
	args = append(args, "-i", url, "-vn", "-f", "s16le", "-ac", "1", "-ar", "44100", "pipe:1")
	cmd := exec.Command(bin, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	m := &Meter{cmd: cmd, done: make(chan struct{}), out: make(chan Sample, 8)}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	go m.read(stdout)
	return m, nil
}

func (m *Meter) Samples() <-chan Sample {
	return m.out
}

func (m *Meter) Stop() {
	m.mu.Lock()
	cmd := m.cmd
	done := m.done
	m.cmd = nil
	m.mu.Unlock()
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
	<-done
}

func (m *Meter) read(r io.Reader) {
	defer close(m.done)
	defer close(m.out)
	analyzer := NewSpectrumAnalyzer(44100, 24, 20, 18000)
	buf := make([]byte, 4096)
	previous := 0.0
	for {
		n, err := io.ReadFull(r, buf)
		if err != nil {
			return
		}
		level := rms(buf[:n])
		sample := Sample{
			Level:     level,
			Transient: math.Max(0, level-previous),
			Bands:     analyzer.Bands(buf[:n]),
			Live:      true,
		}
		previous = level*0.72 + previous*0.28
		select {
		case m.out <- sample:
		default:
		}
	}
}

func rms(buf []byte) float64 {
	total := 0.0
	count := 0
	for i := 0; i+1 < len(buf); i += 2 {
		v := float64(int16(binary.LittleEndian.Uint16(buf[i:]))) / 32768
		total += v * v
		count++
	}
	if count == 0 {
		return 0
	}
	return math.Min(1, math.Sqrt(total/float64(count))*3.2)
}

package audio

type mpvControl interface {
	args() []string
	command(string) error
	close()
}

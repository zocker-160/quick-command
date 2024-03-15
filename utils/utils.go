package utils

import (
	"os"
	"os/exec"

	"github.com/kgadams/go-shellquote"
	"github.com/mattn/go-runewidth"
)

func RemoveLastRune(s string) string {
	return RemoveNRunes(s, 1)
}

func RemoveNRunes(s string, n int) string {
	l := runewidth.StringWidth(s)
	if n >= l {
		return ""
	}

	return runewidth.Truncate(s, l-n, "")
}

func FillLine(s string, r rune, width int) string {
	for runewidth.StringWidth(s) <= width {
		s += string(r)
	}
	return s
}

func Clamp(val *int, min, max int) {
	if *val < min {
		*val = min
	}

	if *val > max {
		*val = max
	}
}

func GenerateCommand(cmd string) (*exec.Cmd, error) {
	args, err := shellquote.Split(cmd)
	if err != nil {
		return nil, err
	}

	exe := exec.Command(args[0], args[1:]...)
	exe.Stdin = os.Stdin
	exe.Stdout = os.Stdout
	exe.Stderr = os.Stderr
	return exe, nil
}

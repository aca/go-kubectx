// fzfutil is to use fzf as library
package fzfutil

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

// Based on, https://junegunn.kr/2016/02/using-fzf-in-your-program
func FZF(input func(in io.WriteCloser), opts ...string) ([]string, error) {
	fzf, err := exec.LookPath("fzf")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(fzf, opts...)

	cmd.Stderr = os.Stderr

	in, _ := cmd.StdinPipe()
	go func() {
		input(in)
		in.Close()
	}()

	stdout, err := cmd.Output()
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if !ok {
			return nil, err
		} else {
			code := exitError.ExitCode()
			// EXIT STATUS
			//        0      Normal exit
			//        1      No match
			//        2      Error
			//        130    Interrupted with CTRL-C or ESC
			if code == 1 || code == 130 {
				return nil, nil
			}
		}
	}

	return strings.Split(string(stdout), "\n"), nil
}

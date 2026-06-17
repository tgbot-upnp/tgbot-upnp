//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func showMessage(title, message string) {
	switch runtime.GOOS {
	case "darwin":
		_ = exec.Command("osascript", "-e",
			fmt.Sprintf(`display dialog %q with title %q buttons {"OK"} default button "OK"`,
				message, title),
		).Run()
	default:
		fmt.Fprintf(os.Stderr, "[%s]\n%s\n", title, message)
	}
}

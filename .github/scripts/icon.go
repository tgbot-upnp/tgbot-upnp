//go:build ignore

package main

import (
	"os"
	"github.com/tgbot-upnp/tgbot-upnp/icon"
)

func main() {
	os.WriteFile(os.Args[1], icon.GetIcon(), 0o644)
}

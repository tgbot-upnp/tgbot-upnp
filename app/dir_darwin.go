//go:build darwin

package app

import "os"

func defaultDataDir() string {
	home, _ := os.UserHomeDir()
	return home + "/Library/Application Support/tgbot-upnp"
}

//go:build darwin

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const launchAgentPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>Label</key><string>com.tgbot-upnp.tgbot-upnp</string>
<key>ProgramArguments</key><array><string>%s</string></array>
<key>RunAtLoad</key><true/>
<key>KeepAlive</key><false/>
</dict></plist>`

func launchAgentPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", "com.tgbot-upnp.tgbot-upnp.plist")
}

func isAutostartEnabled() bool {
	_, err := os.Stat(launchAgentPath())
	return err == nil
}

func setAutostart(enable bool) error {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "Library", "LaunchAgents")
	_ = os.MkdirAll(dir, 0o755)

	path := launchAgentPath()
	if enable {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		content := fmt.Sprintf(launchAgentPlist, exe)
		return os.WriteFile(path, []byte(content), 0o644)
	}
	return os.Remove(path)
}

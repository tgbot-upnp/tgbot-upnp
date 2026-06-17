package main

import (
	"os"

	"golang.org/x/sys/windows/registry"
)

const autostartKey = `Software\Microsoft\Windows\CurrentVersion\Run`
const autostartName = "tgbot-upnp"

func isAutostartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, autostartKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(autostartName)
	return err == nil
}

func setAutostart(enable bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, autostartKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if enable {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		return k.SetStringValue(autostartName, exe)
	}
	return k.DeleteValue(autostartName)
}

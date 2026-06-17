package main

import "golang.org/x/sys/windows"

func showMessage(title, message string) {
	msg, _ := windows.UTF16PtrFromString(message)
	cap, _ := windows.UTF16PtrFromString(title)
	_, _ = windows.MessageBox(0, msg, cap, windows.MB_OK|windows.MB_ICONINFORMATION)
}

//go:build !windows && !darwin

package main

func isAutostartEnabled() bool       { return false }
func setAutostart(enable bool) error { return nil }

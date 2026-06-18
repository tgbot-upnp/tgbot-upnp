//go:build !darwin

package app

func defaultDataDir() string { return "." }

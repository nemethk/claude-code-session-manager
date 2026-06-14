package cache

import (
	"os"
	"path/filepath"
	"strings"
)

func dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(home, ".cache", "ccsm")
	return d, os.MkdirAll(d, 0755)
}

func Get(uuid string) (string, bool) {
	d, err := dir()
	if err != nil {
		return "", false
	}
	data, err := os.ReadFile(filepath.Join(d, uuid+".txt"))
	if err != nil {
		return "", false
	}
	s := strings.TrimSpace(string(data))
	return s, s != ""
}

func Set(uuid, summary string) error {
	d, err := dir()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(d, uuid+".txt"), []byte(summary), 0644)
}

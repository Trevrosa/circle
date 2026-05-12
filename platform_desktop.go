//go:build linux || darwin || windows

package main

import "os"

const compactSave = false

func save(data []byte) error {
	return os.WriteFile("people.json", data, 0644)
}

func saveBackup(data []byte) error {
	return os.WriteFile("people.json.bak", data, 0644)
}

func load(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func getInitial() (string, error) {
	file, err := load("people.txt")
	if err != nil {
		return "", err
	}
	return string(file), nil
}

func platformSetup(_ []Person) {}

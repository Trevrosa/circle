//go:build js && wasm

package main

import (
	"errors"
	"fmt"
	"strings"
	"syscall/js"
)

const compactSave = true

func save(data []byte) error {
	js.Global().Get("localStorage").Call("setItem", "people", string(data))
	return nil
}

func saveBackup(_ []byte) error {
	return nil
}

func load(name string) ([]byte, error) {
	name = strings.Split(name, ".")[0]
	item := js.Global().Get("localStorage").Call("getItem", name)
	if item.IsNull() {
		return nil, errors.New("item null")
	}
	return []byte(item.String()), nil
}

func getInitial() (string, error) {
	input := js.Global().Call("prompt", "Enter names separated by commas")
	if input.String() == "" {
		return "", errors.New("no names entered")
	}
	return input.String(), nil
}

func platformSetup(people []Person) {
	closeHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if err := savePeople(people); err != nil {
			fmt.Printf("Error saving people: %v", err)
		}
		return nil
	})
	js.Global().Call("addEventListener", "beforeunload", closeHandler)
}

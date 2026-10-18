package lsp

import (
	"os"
	"path/filepath"
)

func mustWrite(t interface{ Helper() }, path, body string) {
	if err := writeFile(path, body); err != nil {
		panic(err)
	}
}

func mustDir(t interface{ Helper() }, path string) {
	if err := os.MkdirAll(path, 0o777); err != nil {
		panic(err)
	}
}

func writeFile(path, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(body), 0o666)
}

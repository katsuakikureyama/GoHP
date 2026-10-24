package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type artifact struct {
	path string
	data []byte
}

// publishArtifacts stages complete files beside their destinations, then links
// them into place without replacing existing entries. The report is last.
// This is not a multi-file filesystem transaction: consumers must wait for a
// successful CLI exit and verify hashes before using a candidate.
func publishArtifacts(files []artifact) (err error) {
	paths := make([]string, len(files))
	seen := map[string]bool{}
	for i, file := range files {
		abs, e := filepath.Abs(file.path)
		if e != nil {
			return e
		}
		parent, e := filepath.EvalSymlinks(filepath.Dir(abs))
		if e != nil {
			return fmt.Errorf("output directory: %w", e)
		}
		path := filepath.Join(parent, filepath.Base(abs))
		if seen[path] {
			return fmt.Errorf("output and report must have distinct paths")
		}
		seen[path] = true
		if _, e = os.Lstat(path); e == nil {
			return fmt.Errorf("destination already exists: %s (use a fresh candidate path)", file.path)
		} else if !errors.Is(e, os.ErrNotExist) {
			return e
		}
		paths[i] = path
	}

	temps := make([]string, len(files))
	identities := make([]os.FileInfo, len(files))
	published := 0
	defer func() {
		if err != nil {
			for i := published - 1; i >= 0; i-- {
				info, e := os.Lstat(paths[i])
				if errors.Is(e, os.ErrNotExist) {
					continue
				}
				if e != nil {
					err = errors.Join(err, e)
					continue
				}
				// Never remove a replacement installed by another process.
				if !os.SameFile(info, identities[i]) {
					err = errors.Join(err, fmt.Errorf("rollback skipped changed destination %s", paths[i]))
					continue
				}
				err = errors.Join(err, os.Remove(paths[i]))
			}
		}
		for _, temp := range temps {
			if temp != "" {
				err = errors.Join(err, os.Remove(temp))
			}
		}
	}()
	for i, file := range files {
		f, e := os.CreateTemp(filepath.Dir(paths[i]), ".gohp-*")
		if e != nil {
			return e
		}
		temps[i] = f.Name()
		_, e = f.Write(file.data)
		if e == nil {
			e = f.Sync()
		}
		if e == nil {
			identities[i], e = f.Stat()
		}
		if e = errors.Join(e, f.Close()); e != nil {
			return e
		}
	}
	for i, temp := range temps {
		if e := os.Link(temp, paths[i]); e != nil {
			return fmt.Errorf("publish %s: %w", paths[i], e)
		}
		published++
	}
	return nil
}

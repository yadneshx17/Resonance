package library

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Browser struct {
	RootPath    string
	CurrentPath string
	Entries     []Entry
	History     []string
}

func NewBrowser(rootPath string) (*Browser, error) {
	b := &Browser{
		RootPath:    rootPath,
		CurrentPath: rootPath,
	}
	if err := b.ReadDir(); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *Browser) ReadDir() error {
	items, err := os.ReadDir(b.CurrentPath)
	if err != nil {
		return err
	}

	var entries []Entry
	for _, item := range items {
		name := item.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		fullPath := filepath.Join(b.CurrentPath, name)

		if item.IsDir() {
			entries = append(entries, Entry{
				Name:  name,
				Path:  fullPath,
				IsDir: true,
			})
		} else if strings.HasSuffix(strings.ToLower(name), ".mp3") {
			entries = append(entries, Entry{
				Name:  name,
				Path:  fullPath,
				IsDir: false,
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})

	b.Entries = entries
	return nil
}

func (b *Browser) Open(index int) error {
	if index < 0 || index >= len(b.Entries) {
		return nil
	}
	entry := b.Entries[index]
	if !entry.IsDir {
		return nil
	}
	b.History = append(b.History, b.CurrentPath)
	b.CurrentPath = entry.Path
	return b.ReadDir()
}

func (b *Browser) GoBack() error {
	if len(b.History) == 0 {
		return nil
	}
	b.CurrentPath = b.History[len(b.History)-1]
	b.History = b.History[:len(b.History)-1]
	return b.ReadDir()
}

func (b *Browser) CanGoBack() bool {
	return len(b.History) > 0
}

func (b *Browser) CurrentName() string {
	return filepath.Base(b.CurrentPath)
}

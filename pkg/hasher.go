package pkg

import (
	"context"
	"golang.org/x/sync/errgroup"
	"io"
	"io/fs"
	"strings"
	"sync"
	"unicode"
)

type HasherConfigs struct {
	cuncurrentLimit int
}
type Hasher struct {
	storage map[string]map[string]struct{}
	mutex   sync.Mutex
	configs HasherConfigs
}

func (hasher *Hasher) Add(word, fileName string) {
	hasher.mutex.Lock()
	defer hasher.mutex.Unlock()

	fileNamesMap, ok := hasher.storage[word]
	if !ok {
		fileNamesMap = map[string]struct{}{}
		hasher.storage[word] = fileNamesMap
	}
	fileNamesMap[fileName] = struct{}{}
}
func (hasher *Hasher) GetContainFiles(word string) []string {
	hasher.mutex.Lock()
	defer hasher.mutex.Unlock()

	fileNamesMap, ok := hasher.storage[word]
	if ok {
		fileNames := make([]string, 0, len(fileNamesMap))
		for fileName := range fileNamesMap {
			fileNames = append(fileNames, fileName)
		}
		return fileNames
	}
	return nil
}

func (hasher *Hasher) addFiles(f fs.FS) error {
	files, err := FilesFS(f, ".")

	if err != nil {
		return err
	}

	errGr, _ := errgroup.WithContext(context.Background())
	errGr.SetLimit(hasher.configs.cuncurrentLimit)

	for _, fileName := range files {
		errGr.Go(func() error {
			content, err := readFile(f, fileName)
			if err != nil {
				return err
			}

			for _, word := range strings.FieldsFunc(content, isNotLetter) {
				hasher.Add(word, fileName)
			}
			return nil
		})
	}

	if err := errGr.Wait(); err != nil {
		return err
	}

	return nil
}

func isNotLetter(c rune) bool {
	return !unicode.IsLetter(c) && !unicode.IsDigit(c)
}

func readFile(f fs.FS, fileName string) (string, error) {
	open, err := f.Open(fileName)
	if err != nil {
		return "", err
	}
	defer open.Close()

	fileContent, err := io.ReadAll(open)
	if err != nil {
		return "", err
	}
	return string(fileContent), nil
}

package pkg

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"
)

type Searcher struct {
	FS fs.FS
}

func (s *Searcher) Search(word string) (files []string, err error) {
	var hasher = &Hasher{
		storage: map[string]map[string]struct{}{},
		configs: HasherConfigs{
			cuncurrentLimit: 10,
		},
	}
	err = hasher.addFiles(s.FS)
	if err != nil {
		return nil, err
	}

	return hasher.GetContainFiles(word), nil
}

func (s *Searcher) RequestHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/files/search", s.handleSearch)
	return mux
}

func (searcher *Searcher) handleSearch(w http.ResponseWriter, r *http.Request) {
	wordToSearch := strings.TrimSpace(r.URL.Query().Get("text"))
	if wordToSearch == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fileNames, err := searcher.Search(wordToSearch)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(fileNames)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

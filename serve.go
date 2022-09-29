package main

import (
	"bytes"
	_ "embed"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed placeholder.gif
var placeholderGif []byte

//go:embed placeholder.png
var placeholderPng []byte

//go:embed placeholder.jpg
var placeholderJpg []byte

func serve(po *processOptions, addr string) {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		uri := r.RequestURI
		if strings.HasSuffix(uri, "/") {
			uri = uri + "index"
		}

		if strings.Contains(uri, "..") {
			http.Error(w, "Invalid path", http.StatusNotFound)
			return
		} else if strings.HasPrefix(uri, ".") {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		} else if !strings.HasPrefix(uri, "/") {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		// ResolvePath

		ext := filepath.Ext(uri)

		if ext != "" {
			var fileBytes []byte
			f, err := po.ResolvePath(uri)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				log.Println(err.Error())
				w.WriteHeader(http.StatusNotFound)
				return
			} else if err != nil && errors.Is(err, os.ErrNotExist) {
				// TODO add in something to decide whether to make placeholders or not
				if ext == ".gif" {
					fileBytes = placeholderGif
				} else if ext == ".jpg" || ext == ".jpeg" {
					fileBytes = placeholderJpg
				} else if ext == ".png" {
					fileBytes = placeholderPng
				}

				if fileBytes == nil || !po.UsePlaceholderImages {
					log.Println(err.Error())
					w.WriteHeader(http.StatusNotFound)
					return
				}

			} else if fileBytes, err = ioutil.ReadFile(f); err != nil {
				log.Println(err.Error())
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			w.Write(fileBytes)

		} else if ext == "" {
			path := filepath.Join(po.Source, uri+po.Extension)

			root, err := parseFile(path)
			if err != nil {
				http.Error(w, "Internal Error", http.StatusInternalServerError)
				w.Write([]byte("<h1>Compilation Failed</h1><p>" + err.Error() + "</p>"))
				return
			}

			nodeState := &Scope{
				FileScope: &FileScope{
					Options: po,
				},
			}
			out := &bytes.Buffer{}
			if err = formatRoot(nodeState, root, out); err != nil {
				http.Error(w, "Internal Error", http.StatusInternalServerError)
				w.Write([]byte("<h1>Compilation Failed</h1><p>" + err.Error() + "</p>"))
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			w.Write(out.Bytes())
		}
	})

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

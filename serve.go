package main

import (
	"bytes"
	_ "embed"
	"errors"
	"io/ioutil"
	"log"
	"mime"
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

	err2 := mime.AddExtensionType(".css", "text/css")
	if err2 != nil {
		log.Printf("Error in mime js %s", err2.Error())
	}

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

		gs := NewGlobalScope()


		// ResolvePath

		ext := filepath.Ext(uri)

		if ext == "" {
		} else if ext != ".html" {
			var fileBytes []byte
			f, err := po.ResolvePath(uri)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				log.Println(err.Error())
				w.WriteHeader(http.StatusNotFound)
				return
			} else if err != nil && errors.Is(err, os.ErrNotExist) {

				if po.UsePlaceholderImages {
					if ext == ".gif" {
						fileBytes = placeholderGif
					} else if ext == ".jpg" || ext == ".jpeg" {
						fileBytes = placeholderJpg
					} else if ext == ".png" {
						fileBytes = placeholderPng
					}
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


			//mime.AddExtensionType(".css", "text/css; charset=utf-8")

			log.Println("Serving " + uri + " extension: " + ext)
			if ext == ".css" {
				w.Header().Set("Content-Type", "text/css; charset=utf-8")
			} else {
				w.Header().Set("Content Type", "application/octet-stream")
			}
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			w.Header().Write(log.Writer())
			w.WriteHeader(http.StatusOK) // must be last to force our Content-Type to be respected
			w.Write(fileBytes)

		} else if ext == ".html" && filepath.Base(ext)[0] != '_' {
			path := filepath.Join(po.Source, uri[0:len(uri)-len(ext)]+po.Extension)

			root, err := parseFile(path)
			if err != nil {
				http.Error(w, "Internal Error", http.StatusInternalServerError)
				w.Write([]byte("<h1>Compilation Failed</h1><p>" + err.Error() + "</p>"))
				return
			}

			nodeState := &Scope{
				FileScope: &FileScope{
					Path: path,
					GlobalScope: gs,
					Options: po,
					UniqueClass: &HtmlRenderingBuffer{},
				},
			}
			out := &bytes.Buffer{}
			if err = formatRoot(nodeState, root, out); err != nil {
				http.Error(w, "Internal Error", http.StatusInternalServerError)
				w.Write([]byte("<h1>Compilation Failed</h1><p>" + err.Error() + "</p>"))
				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			w.WriteHeader(http.StatusOK)
			w.Write(out.Bytes())
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

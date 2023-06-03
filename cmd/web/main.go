package main

import (
	"log"
	"net/http"
	"path/filepath"
)

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", home)
	//Calculator
	mux.HandleFunc("/calculator", calculator)
	mux.HandleFunc("/formHandler", formHandler)
	mux.HandleFunc("/deleteEntry", deleteEntry)
	//ToDoList
	mux.HandleFunc("/todolist", toDoList)
	mux.HandleFunc("/formHandlerToDoList", formHandlerToDoList)
	mux.HandleFunc("/deleteEntryToDoList", deleteEntryToDoList)
	mux.HandleFunc("/editPost", editPost)

	fileServer := http.FileServer(neuteredFileSystem{http.Dir("../../ui/static/")})
	mux.Handle("/static", http.NotFoundHandler())
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	log.Println("Сервер запушен на : 4000 порту")
	err := http.ListenAndServe(":4000", mux)
	log.Fatalln(err)
}

type neuteredFileSystem struct {
	fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}

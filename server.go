package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"html"
	"log"
	"net/http"
	"os/exec"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/write", Write)
	router.HandleFunc("/query", Query)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func Write(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	out, err := exec.Command("go", "run", "utils.go", "write.go", "--data", r.URL.Query()["data"][0]).CombinedOutput()
	if err == nil {
		fmt.Fprintf(w, "1")
	} else {
		fmt.Fprintf(w, err.Error()+string(out))
	}
}

func Query(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	params := r.URL.Query()
	args := []string{"run", "comparison.go", "scanner.go", "parser.go", "token.go", "utils.go", "query.go"}
	if query, ok := params["query"]; ok {
		args = append(args, "--query", query[0])
	} else {
		fmt.Fprintf(w, "'query' param required")
		return
	}
	if start, ok := params["start"]; ok {
		args = append(args, "--start", start[0])
	}
	if end, ok := params["end"]; ok {
		args = append(args, "--end", end[0])
	}
	out, err := exec.Command("go", args...).CombinedOutput()
	if err == nil {
		fmt.Fprintf(w, string(out))
	} else {
		fmt.Fprintf(w, err.Error()+string(out))
	}
}

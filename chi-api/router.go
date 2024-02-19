package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	//Default root response
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	//Api to server emails
	r.Get("/api/enron", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Query())
		records, err := queryEnron(r.URL.Query())
		if err != nil {
			log.Print(err)
		}
		w.Write(records)
	})

	//Static File Serving
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "..", "enron_gui", "vis_enron", "dist"))
	FileServer(r, "/app", filesDir)

	http.ListenAndServe(":3000", r)
}

// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

func queryEnron(queryStr url.Values) ([]byte, error) {
	// ct := time.Now()

	num, err := strconv.Atoi(queryStr["n_from"][0])
	if err != nil {
		// Handle error if the string is not a valid integer
		fmt.Println("Error:", err)
	}
	results, err := strconv.Atoi(queryStr["max_results"][0])
	if err != nil {
		// Handle error if the string is not a valid integer
		fmt.Println("Error:", err)
	}

	query := map[string]interface{}{
		"search_type": "querystring",
		"query":       map[string]interface{}{"term": queryStr["query"][0]},

		"from":        (num),
		"max_results": results,
		"_source":     []string{},
	}

	jason, err := json.Marshal(query)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(jason))

	url := "http://localhost:4080/api/enron_mail/_search"
	// req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(json)))

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jason)))

	if err != nil {
		return nil, err
	}
	req.Header.Set("Content_Type", "application/json")
	req.SetBasicAuth("admin", "Complexpass#123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return io.ReadAll(resp.Body)
	}

	return nil, fmt.Errorf("error requesting results: %s", resp.Status)

}

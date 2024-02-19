package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
)

type Filec struct {
	content string
}

var cpuprofile = flag.String("cpuprofile", "cpu.prof", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "mem.prof", "write memory profile to `file`")

func main() {

	MAX_ITERS := flag.Int("iters", int(math.Inf(1)), "Max documents to upload")

	flag.Parse()

	INDEX_NAME := flag.Arg(0)

	//SUPPORT FOR PROFILING
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	// fmt.Println(doc_json)
	// files, err := FilePathWalkDir("enron_mail_20110402")
	files, err := FilePathWalkDir("enron_mail_20110402/maildir/dorland-c")

	if err != nil {
		fmt.Println(err)
	}

	for i, file := range files {

		if i == *MAX_ITERS {
			fmt.Println("Reached Max Number Of Docs")
			break
		}
		fmt.Printf("Processing File %s, i: %d\r\n", file, i)

		// _, err := indexergo.File2Json(file)
		file_content, err := os.ReadFile(file)

		if err != nil {
			fmt.Println(file)
			fmt.Print(err)
		}

		// var file_struct = map[string]interface{}{"content": "This is just some random content to test 3"}
		var file_struct = map[string]interface{}{"content": string(file_content)}

		json_str, err := json.Marshal(file_struct)

		if err != nil {
			log.SetPrefix(fmt.Sprintf("Proccessing file %s", file))
			log.Panicln(err)
		}

		err = upload_doc(json_str, INDEX_NAME)
		if err != nil {
			log.SetPrefix(fmt.Sprintf("Uploading file %s", file))
			log.Panicln(err)
		}

		// SUPPORT FOR MEM PROFILE
		if *memprofile != "" {
			f, err := os.Create(*memprofile)
			if err != nil {
				log.Fatal("could not create memory profile: ", err)
			}
			defer f.Close() // error handling omitted for example
			runtime.GC()    // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatal("could not write memory profile: ", err)
			}
		}

	}
}

func FilePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func upload_doc(json []byte, index_name string) error {
	url := fmt.Sprintf("http://localhost:4080/api/%s/_doc", index_name)
	// req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(json)))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("admin", "Complexpass#123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Doc Uploaded Successfully!")
		return nil
	}

	return errors.New("Error en el request")
}

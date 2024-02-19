package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"slices"
	"strings"
)

// https://github.com/zincsearch/zincsearch/issues/413 Arbitrary value found here (99% to have margin for additional Json symbols)
const ZINC_SEARCH_MAX_LINE = 1000000 * 0.99

var TOTAL_SINGLE_UPS = 0

type Mail struct {
	Message_ID                string
	Date                      string
	From                      string
	To                        string
	Subject                   string
	Cc                        string
	Mime_Version              string
	Content_Type              string
	Content_Transfer_Encoding string
	Bcc                       string
	X_From                    string
	X_To                      string
	X_cc                      string
	X_bcc                     string
	X_Folder                  string
	X_Origin                  string
	X_FileName                string
	Body                      string
}

func main() {
	//SUPPORT FOR CPU PROFILING
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
	var maildir = flag.String("d", "../enron_mail_20110402/maildir", "Maildir Location")

	flag.Parse()

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

	// var MAX_ITERS = flag.Int("iters", int(math.Inf(1)), "Max documents to upload")

	// var my_mail, err = mail_from_file("../enron_mail_20110402/maildir/weldon-c/kcs/3.")

	// my_mail_json, err := json.Marshal(my_mail)

	// fmt.Println(string(my_mail_json))

	// fmt.Println(len(my_mail_json))

	// return

	INDEX_NAME := flag.Arg(0)

	if INDEX_NAME == "" {
		log.Fatal(errors.New("An Index Name is required"))
	}

	index_json, err := index_json(INDEX_NAME)
	// fmt.Println(string(index_json))

	if err != nil {
		log.Panic(err)
	}

	// //Make a bulk Upload for each person inside maildir
	person_dirs, err := os.ReadDir(*maildir)

	for _, dir := range person_dirs {

		err := upload_person(*maildir, dir.Name(), index_json, INDEX_NAME)

		if err != nil {
			log.SetPrefix(fmt.Sprintf("Error uploading %s\r\n", dir.Name()))
			log.Println(err)
			// log.Panic(err)
		} else {
			log.Print(fmt.Sprintf("Person %s uploaded", dir.Name()))
		}

	}
	fmt.Printf("Total Single Uploads in execution (files bigger than 1MB) %d", TOTAL_SINGLE_UPS)

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

func mail_from_file(file string) (Mail, error) {

	readFile, err := os.Open(file)

	if err != nil {
		return Mail{}, err
	}

	// full_str := fileScanner.Text()
	// fmt.Println(full_str)

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	mail_info := make(map[string]string)

	file_headers := []string{
		"Message-ID", "Date", "From", "To", "Subject", "Cc", "Mime-Version", "Content-Type", "Content-Transfer-Encoding", "Bcc",
		"X-From", "X-To", "X-cc", "X-bcc", "X-Folder", "X-Origin", "X-FileName",
	}

	var current_var, current_val string
	reached_body := false

	for fileScanner.Scan() {
		line := fileScanner.Text()

		//No need to keep checking for headers, body can be the biggest portion of an email
		if reached_body {
			current_val = strings.Join([]string{current_val, line}, "\n")
			continue
		}

		//Beginning of content is marked by an empty line
		if len(line) == 0 {
			current_var = "Body"
			current_val = ""
			reached_body = true
			continue
		}

		line_split := strings.Split(line, ":")
		pref, cont := line_split[0], strings.Join(line_split[1:], ":")

		// If the line begins with one of the headers, I store a separate field, otherwise I keep Storing values in the same field
		if slices.Contains(file_headers, pref) {
			mail_info[current_var] = current_val
			current_var = pref
			current_val = cont
		} else {
			current_val = strings.Join([]string{current_val, line}, "\n")
		}

	}
	mail_info[current_var] = current_val

	readFile.Close()

	mail := Mail{
		Message_ID:                mail_info["Message-ID"],
		Date:                      mail_info["Date"],
		From:                      mail_info["From"],
		To:                        mail_info["To"],
		Subject:                   mail_info["Subject"],
		Cc:                        mail_info["Cc"],
		Mime_Version:              mail_info["Mime-Version"],
		Content_Type:              mail_info["Content_Type"],
		Content_Transfer_Encoding: mail_info["Content_Transfer-Encoding"],
		Bcc:                       mail_info["Bcc"],
		X_From:                    mail_info["X-From"],
		X_To:                      mail_info["X-To"],
		X_cc:                      mail_info["X-cc"],
		X_bcc:                     mail_info["X-bcc"],
		X_Folder:                  mail_info["X-Folder"],
		X_Origin:                  mail_info["X-Origin"],
		X_FileName:                mail_info["X-FileName"],
		Body:                      mail_info["Body"],
	}

	return mail, nil
}

// Function to build the JSON of all the files of a single person
func upload_person(maildir string, person string, index_json []byte, index_name string) error {
	// fmt.Printf("Bulk Upload For Person %s\r\n", person)
	person_dir := fmt.Sprintf("%s/%s", maildir, person)
	// fmt.Printf("Person dir %s\r\n", person_dir)

	files, err := filePathWalkDir(person_dir)

	var bulk_json = bytes.NewBuffer(make([]byte, 0))

	if err != nil {
		return err
	}

	for _, file := range files {

		mail, err := mail_from_file(file)

		fileI, err := os.Stat(file)
		tooBig4Bulk := fileI.Size() >= ZINC_SEARCH_MAX_LINE

		fc_json, err := json.Marshal(mail)
		if err != nil {
			return err
		}

		if tooBig4Bulk {
			var single_json = bytes.NewBuffer(make([]byte, 0))

			single_json.Write(fc_json)

			err = upload_doc(*single_json, index_name)

			if err != nil {
				log.SetPrefix(fmt.Sprintf("Error in single Upload of %s", file))
				log.Println(err)
			}
			continue
		}

		if err != nil {
			return err
		}

		bulk_json.Write(index_json)
		bulk_json.WriteString("\n")

		bulk_json.Write(fc_json)
		bulk_json.WriteString("\n")
	}
	err = upload_bulk(*bulk_json)

	return err

}

// Function to build the json that must precede every content json in the bulk upload
func index_json(index_name string) ([]byte, error) {
	index_map := map[string]map[string]string{"index": {"_index": index_name}}

	return json.Marshal(index_map)
}

func filePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func upload_bulk(json bytes.Buffer) error {
	url := "http://localhost:4080/api/_bulk"
	// req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(json)))

	req, err := http.NewRequest("POST", url, &json)
	req.Header.Set("Content_Type", "application/json")
	req.SetBasicAuth("admin", "Complexpass#123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// fmt.Println(resp)

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return errors.New(fmt.Sprintf("Error Uploading Bulk: %s\r\n", resp.Status))
}

func upload_doc(json bytes.Buffer, index_name string) error {
	url := fmt.Sprintf("http://localhost:4080/api/%s/_doc", index_name)
	// req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(json)))
	req, err := http.NewRequest("POST", url, &json)
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("admin", "Complexpass#123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		TOTAL_SINGLE_UPS++
		return nil
	}

	return errors.New("Error loading single doc")
}

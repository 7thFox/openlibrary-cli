package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

const (
	colorNormal = "\033[0m"

	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

func colorPrintln(color string, msg ...interface{}) {
	if GetSettings().Color {
		defer log.Print(colorNormal)
		log.Print(colorRed)
	}

	log.Println(msg...)
}

func main() {
	log.SetFlags(0)

	if GetSettings().Quiet {
		log.SetOutput(ioutil.Discard)
	}

	lookups := make(chan string)
	lookupResponses := make(chan *BookInfo)

	/*
		in, _ := os.Open("test.txt")
		startReader(in, lookups)
		/*/
	startReader(os.Stdin, lookups)
	//*/
	startURLLookup(lookups, lookupResponses)

	format := CompileFormat(GetSettings().Format)
	for x := range lookupResponses {
		fmt.Println(format.Format(x))
	}
}

func startURLLookup(lookups <-chan string, responses chan<- *BookInfo) {
	defLookupKind := GetSettings().DefaultLookupKind
	nWorkers := GetSettings().NWorkers
	var wgClose sync.WaitGroup

	for i := 0; i < nWorkers; i++ {
		wgClose.Add(1)
		go func() {
			defer wgClose.Done()
			for lookup := range lookups {
				// TODO: one-line lookup kind
				url := fmt.Sprintf("http://openlibrary.org/api/books?bibkeys=%s:%s&format=json&jscmd=data", defLookupKind, lookup)
				resp, err := http.Get(url)
				if err != nil {
					colorPrintln(colorRed, "Error: ", err.Error())
					continue
				}

				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					colorPrintln(colorRed, "Error getting body: ", err.Error())
					continue
				}

				if resp.StatusCode < 200 || resp.StatusCode > 299 {
					colorPrintln(colorRed, "Error Response: ", body)
					continue
				}

				var data map[string]BookInfo
				if err := json.Unmarshal(body, &data); err != nil {
					bodyStr := string(body)
					_ = bodyStr
					colorPrintln(colorRed, "Error parsing Json: ", err.Error())
					continue
				}

				for _, v := range data {
					responses <- &v
				}
			}
		}()
	}
	go func() {
		wgClose.Wait()
		close(responses)
	}()
}

func startReader(in io.Reader, lookups chan<- string) {
	go func() {
		re, _ := regexp.Compile(`^([\d\-]+).*`)
		defer close(lookups)

		scanner := bufio.NewScanner(in)
		for scanner.Scan() {
			// TODO: Eventually need actual parsing
			m := re.FindStringSubmatch(scanner.Text())
			if len(m) >= 2 && m[1] != "" {
				lookups <- strings.ReplaceAll(m[1], "-", "")
			}
		}
		if err := scanner.Err(); err != nil {
			colorPrintln(colorRed, err.Error())
			os.Exit(1)
		}
	}()
}

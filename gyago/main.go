package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	defaultEndpoint := os.Getenv("GYAGO_SERVER")
	if defaultEndpoint == "" {
		defaultEndpoint = "http://gyazo.com/upload.cgi"
	}
	endpoint := flag.String("e", defaultEndpoint, "endpoint to upload")
	authenticate := flag.String("a", os.Getenv("GYAGO_BASICAUTH"), "basic authentication")
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	// get hostname for filename
	url_, err := url.Parse(*endpoint)
	if err != nil {
		log.Fatalf("url.Parse: %v", err)
	}
	host, _, err := net.SplitHostPort(url_.Host)
	if err != nil {
		host = url_.Host
	}

	// make content
	content, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatalf("ReadFile: %v", err)
	}

	// %Y%m%d%H%M%S
	id := time.Now().Format("20060102030405")

	// create multipart
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	err = w.WriteField("id", id)
	if err != nil {
		log.Fatalf("WriteField: %v", err)
	}
	part, err := w.CreateFormFile("imagedata", host)
	if err != nil {
		log.Fatalf("CreateFormFile: %v", err)
	}
	part.Write(content)
	err = w.Close()
	if err != nil {
		log.Fatalf("Close: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, *endpoint, &b)
	if err != nil {
		log.Fatalf("NewRequest: %v", err)
	}
	if *authenticate != "" {
		if token := strings.SplitN(*authenticate, ":", 2); len(token) == 2 {
			req.SetBasicAuth(token[0], token[1])
		}
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	// then, upload
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Post: %v", err)
	}
	defer res.Body.Close()

	content, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("ReadAll: %v", err)
	}
	fmt.Println(string(content))
}

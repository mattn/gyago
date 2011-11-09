package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"mime/multipart"
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
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	// get hostname for filename
	url_, err := url.Parse(*endpoint)
	if err != nil {
		log.Fatalf("ReadFile: %v", err)
	}
	host := strings.SplitN(url_.Host, ":", 2)[0]

	// make content
	content, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatalf("ReadFile: %v", err)
	}

	// %Y%m%d%H%M%S
	id := time.LocalTime().Format("20060102030405")

	// create multipart
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	err = w.WriteField("id", id)
	part, err := w.CreateFormFile("imagedata", host)
	if err != nil {
		log.Fatalf("CreateFormFile: %v", err)
	}
	part.Write(content)
	err = w.Close()
	if err != nil {
		log.Fatalf("Close: %v", err)
	}
	body := strings.NewReader(b.String())

	// then, upload
	res, err := http.Post(*endpoint, w.FormDataContentType(), body)
	if err != nil {
		log.Fatalf("Post: %v", err)
	}
	defer res.Body.Close()

	content, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("ReadAll: %v", err)
	}
	println(string(content))
}

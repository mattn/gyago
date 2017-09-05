package gyazo

import (
	"compress/gzip"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
	"time"

	"appengine"
	"appengine/datastore"
)

type Gyazo struct {
	Created time.Time
	Data    []byte
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func TopPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf8")
	b, _ := ioutil.ReadFile("index.html")
	w.Write(b)
}

func Image(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	_, id := path.Split(r.URL.Path)
	ext := path.Ext(id)
	id = id[:len(id)-len(ext)]
	gyazo := &Gyazo{}
	key := datastore.NewKey(c, "Gyazo", id, 0, nil)
	if err := datastore.Get(c, key, gyazo); err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf8")
		http.Error(w, err.Error(), 500)
	} else {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("ETag", id)
		w.Write(gyazo.Data)
	}
}

func Upload(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if r.Method != "POST" {
		http.Error(w, "invalid request", 500)
		return
	}
	ct := r.Header.Get("Content-Type")
	if strings.SplitN(ct, ";", 2)[0] != "multipart/form-data" {
		http.Error(w, "invalid request", 500)
		return
	}
	_, params, err := mime.ParseMediaType(ct)
	if err != nil {
		http.Error(w, "invalid request", 500)
		return
	}
	boundary, ok := params["boundary"]
	if !ok {
		http.Error(w, "invalid request", 500)
		return
	}
	reader := multipart.NewReader(r.Body, boundary)
	var image []byte
	for {
		part, err := reader.NextPart()
		if part == nil || err != nil {
			break
		}
		if part.FormName() != "imagedata" {
			continue
		}
		v := part.Header.Get("Content-Disposition")
		if v == "" {
			continue
		}
		d, _, err := mime.ParseMediaType(v)
		if err != nil {
			continue
		}
		if d != "form-data" {
			continue
		}
		image, _ = ioutil.ReadAll(part)
	}
	gyazo := &Gyazo{
		Created: time.Now(),
		Data:    image,
	}

	sha := sha1.New()
	sha.Write(image)
	id := fmt.Sprintf("%x", string(sha.Sum(nil))[0:8])
	key := datastore.NewKey(c, "Gyazo", id, 0, nil)
	_, err = datastore.Put(c, key, gyazo)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	host := r.Host
	hi := strings.SplitN(host, ":", 2)
	if len(hi) == 2 && hi[1] == "80" {
		host = hi[0]
	}
	w.Write([]byte("https://" + host + "/" + id + ".png"))
}

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()
			w = gzipResponseWriter{Writer: gz, ResponseWriter: w}
		}

		if r.URL.Path == "/" {
			if r.Method == "POST" {
				Upload(w, r)
			} else {
				TopPage(w, r)
			}
		} else {
			Image(w, r)
		}
	})
}

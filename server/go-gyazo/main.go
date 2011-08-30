package gyazo

import (
	"appengine"
	"appengine/datastore"
	"fmt"
	"http"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"path"
	"crypto/sha1"
	"strings"
	"time"
)

type Gyazo struct {
	Created datastore.Time
	Data    []byte
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
	key := datastore.NewKey("Gyazo", id, 0, nil)
	if err := datastore.Get(c, key, gyazo); err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf8")
		http.Error(w, err.String(), 500)
	} else {
		w.Header().Set("Content-Type", "image/png")
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
	if strings.Split(ct, ";", 2)[0] != "multipart/form-data" {
		http.Error(w, "invalid request", 500)
		return
	}
	_, params := mime.ParseMediaType(ct)
	boundary, ok := params["boundary"]
	if !ok {
		http.Error(w, "invalid request", 500)
		return
	}
	reader := multipart.NewReader(r.Body, boundary)
	var image []byte
	for {
		part, err := reader.NextPart()
		if part == nil {
			break
		}
		data, _ := ioutil.ReadAll(part)
		v := part.Header.Get("Content-Disposition")
		if v == "" {
			continue
		}
		d, params := mime.ParseMediaType(v)
		if d != "form-data" {
			continue
		}
		if params["filename"] != "" {
			image = data
		}

	}
	gyazo := &Gyazo{
		Created: datastore.SecondsToTime(time.Seconds()),
		Data:    image,
	}

	sha := sha1.New()
	sha.Write(image)
	id := fmt.Sprintf("%x", string(sha.Sum())[0:8])
	key := datastore.NewKey("Gyazo", id, 0, nil)
	_, err := datastore.Put(c, key, gyazo)
	if err != nil {
		http.Error(w, err.String(), 500)
		return
	}
	host := r.Host
	hi := strings.Split(host, ":", 2)
	if len(hi) == 2 && hi[1] == "80" {
		host = hi[0]
	}
	w.Write([]byte("http://" + host + "/" + id + ".png"))
}

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

package gyazo

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

// Gyazo is struct of entity of gyazo.
type Gyazo struct {
	Created time.Time
	Data    []byte
}

func serveImage(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	_, id := path.Split(r.URL.Path)
	ext := path.Ext(id)
	id = id[:len(id)-len(ext)]
	if id == r.Header.Get("If-None-Match") {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	gyazo := &Gyazo{}
	item, err := memcache.Get(c, id)
	if err == nil {
		gyazo.Data = item.Value
	} else {
		key := datastore.NewKey(c, "Gyazo", id, 0, nil)
		if err := datastore.Get(c, key, gyazo); err != nil {
			if err == datastore.ErrNoSuchEntity {
				log.Infof(c, "serveImage: %v", err)
				http.Error(w, http.StatusText(http.StatusNotFound), 404)
			} else {
				log.Criticalf(c, "serveImage: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}
		memcache.Set(c, &memcache.Item{
			Key:   id,
			Value: gyazo.Data,
		})
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("ETag", id)
	w.Write(gyazo.Data)
}

func uploadImage(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if r.URL.Path != "/" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	f, h, err := r.FormFile("imagedata")
	if err != nil {
		log.Criticalf(c, "uploadImage: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer f.Close()
	if ct := h.Header.Get("Content-Type"); ct != "image/png" && ct != "application/octet-stream" {
		log.Warningf(c, "content-type should be image/png: %v", ct)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Criticalf(c, "uploadImage: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	gyazo := &Gyazo{
		Created: time.Now(),
		Data:    b,
	}

	sum := sha1.Sum(b)
	id := fmt.Sprintf("%x", string(sum[:])[0:8])

	key := datastore.NewKey(c, "Gyazo", id, 0, nil)
	_, err = datastore.Put(c, key, gyazo)
	if err != nil {
		log.Criticalf(c, "uploadImage: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf8")
	w.Write([]byte("https://" + r.Host + "/" + id + ".png"))
}

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			uploadImage(w, r)
		case http.MethodGet:
			if r.URL.Path == "/" {
				http.ServeFile(w, r, "index.html")
			} else {
				serveImage(w, r)
			}
		}
	})
}

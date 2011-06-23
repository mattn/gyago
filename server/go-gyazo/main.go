package gyazo

import (
	"http"
	"os"
	"log"
	"io/ioutil"
	"github.com/hoisie/web.go"
)

func init() {
	s := &web.Server{
		Config: web.Config,
		Logger: log.New(os.Stderr, "", log.LstdFlags),
	}

	s.Get("/", func(ctx *web.Context) {
		b, _ := ioutil.ReadFile("index.html")
		ctx.Write(b)
	})
	s.Post("/upload.cgi", func(ctx *web.Context) {
		for k, f := range ctx.Request.Files {
			s.Logger.Println(k)
			if k == "imagedata" {
				ctx.Write([]byte("done"))
				break
			}
		}
	})

	http.Handle("/", s)
}

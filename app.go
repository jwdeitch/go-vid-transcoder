package main

import (
	"fmt"
	"net/http"
	"time"
	"github.com/emirozer/go-helpers"
)

type Visitor struct {
	RemoteAddr   string
	ForwardedFor string
	Time         int64
	UserAgent    string
}

type Visit struct {
	Visitor           Visitor
	RequestedResource string
	Method            string
}

func main() {
	fmt.Println("We're up and running")

	http.HandleFunc("/video", getVideo) // GET video

	err := http.ListenAndServe(":9090", nil)
	helpers.Check(err)

}

func getVideo(w http.ResponseWriter, r *http.Request) {
	hostname := r.URL.Query().Get("id")
	logVisit(r)
	fmt.Fprintf(w, hostname)
}

func logVisit(r *http.Request) {
	visitor := Visit{
		Visitor: Visitor{
			RemoteAddr:r.RemoteAddr,
			ForwardedFor:r.Header.Get("X-FORWARDED-FOR"),
			Time:time.Now().Unix(),
			UserAgent:r.UserAgent(),
		},
		RequestedResource: r.URL.Path,
		Method: r.Method,
	}
	fmt.Println(visitor)
	// TODO: Log
}

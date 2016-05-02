package main

import (
	"fmt"
	"net/http"
	"github.com/emirozer/go-helpers"
)

func main() {
	fmt.Println("We're up and running")

	http.HandleFunc("/video", getVideo) // GET video

	err := http.ListenAndServe(":9090", nil)
	helpers.Check(err)

}

func getVideo(w http.ResponseWriter, r *http.Request) {
	hostname := r.URL.Query().Get("id")
	fmt.Fprintf(w, hostname)
}

package main

import (
	"fmt"
	"net/http"
	"time"
	"os"
	"io"
	"github.com/emirozer/go-helpers"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
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

type Object struct {
	Name string
	Size int64
	LastModified string
}

func main() {
	fmt.Println("We're up and running")

	http.HandleFunc("/",index)
	http.HandleFunc("/video", getVideo) // GET video
	http.HandleFunc("/video/upload", uploadVideo) // POST upload video

	err := http.ListenAndServe(":9090", nil)
	helpers.Check(err)

}

func index(w http.ResponseWriter, r *http.Request) {
	logVisit(r)
	creds := credentials.NewEnvCredentials()

	svc := s3.New(session.New(), &aws.Config{
		Region: aws.String("us-west-2"),
		Credentials: creds})

	lsObjs := s3.ListObjectsInput{
		Bucket: aws.String("transcoderoutput489349"),
		Prefix: aws.String("videos")}

	objects, err := svc.ListObjects(&lsObjs)
	helpers.Check(err)



	fmt.Fprintf(w, objects.String())
	fmt.Fprintf(w, "index")
}

//Thanks! http://sanatgersappa.blogspot.com/2013/03/handling-multiple-file-uploads-in-go.html
func uploadVideo(w http.ResponseWriter, r *http.Request) {
	logVisit(r)
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		//parse the multipart form in the request
		err := r.ParseMultipartForm(100000)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//get a ref to the parsed multipart form
		m := r.MultipartForm

		//get the *fileheaders
		files := m.File["myfiles"]
		for i, _ := range files {
			//for each fileheader, get a handle to the actual file
			file, err := files[i].Open()
			defer file.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//create destination file making sure the path is writeable.
			dst, err := os.Create("/tmp/" + files[i].Filename)
			defer dst.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//copy the uploaded file to the destination file
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		}
		fmt.Fprintf(w, "Upload successful.")
	}

}

func getVideo(w http.ResponseWriter, r *http.Request) {
	logVisit(r)
	hostname := r.URL.Query().Get("id")
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

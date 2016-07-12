package main

import (
	"fmt"
	"net/http"
	"time"
	"os"
	"encoding/json"
	"io"
	"regexp"
	"sort"
	"github.com/inturn/go-helpers"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"strings"
)

const bucket string = "transcoderoutput489349"
const region string = "us-west-2"

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

type S3Object struct {
	Name         string
	Size         int64
	LastModified int64
	thumbCount   int
}

type objectListItem []S3Object

func main() {
	fmt.Println("We're up and running on port 9090")

	http.HandleFunc("/", index)
	http.HandleFunc("/video", getVideo) // GET video
	http.HandleFunc("/video/upload", uploadVideo) // POST upload video

	err := http.ListenAndServe(":9090", nil)
	helpers.Check(err)

}

func (s objectListItem) Len() int {
	return len(s)
}
func (s objectListItem) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s objectListItem) Less(i, j int) bool {
	return s[i].LastModified < s[j].LastModified
}

func index(w http.ResponseWriter, r *http.Request) {
	logVisit(r)
	creds := credentials.NewEnvCredentials()

	svc := s3.New(session.New(), &aws.Config{
		Region: aws.String(region),
		Credentials: creds})

	lsObjs := s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String("videos/output")}

	objects, err := svc.ListObjects(&lsObjs)
	helpers.Check(err)

	var S3Objlist objectListItem

	for _, s3Item := range objects.Contents {

		thumbnailCount := 0;

		// A key with a '/' as the last char is a directory
		if helpers.LastNCharacters(*s3Item.Key, 1) == "/" {
			continue
		}

		// Will will count the thumbnails here
		r, _ := regexp.Compile("([^\\/]*)-");
		// some nasty regex here
		match := strings.TrimSuffix(r.FindString(*s3Item.Key), "-")
		// since we can't itemize by file name
		if (match != "") {
			thumbnailCount++
		} else {
			lastModified := *s3Item.LastModified
			S3Objlist = append(S3Objlist, S3Object{
				*s3Item.Key, *s3Item.Size, lastModified.Unix(), thumbnailCount})
		}
	}

	sort.Sort(S3Objlist)

	response, _ := json.Marshal(S3Objlist)

	fmt.Fprintf(w, string(response))
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
		files := m.File["files"]

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

package main

import (
	"encoding/json"
	"regexp"
	"sort"
	"github.com/inturn/go-helpers"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"strings"

	"github.com/apex/go-apex"
	"io/ioutil"
	"os"
)

type message struct {
	Value string `json:"value"`
}

type S3Object struct {
	Name         string
	Size         int64
	LastModified int64
	thumbCount   int
}

type Env struct {
	AWSACCESSKEYID string `json:"AWS_ACCESS_KEY_ID"`
	AWSSECRETACCESSKEY string `json:"AWS_SECRET_ACCESS_KEY"`
}

type objectListItem []S3Object

func (s objectListItem) Len() int {
	return len(s)
}
func (s objectListItem) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s objectListItem) Less(i, j int) bool {
	return s[i].LastModified < s[j].LastModified
}

func main() {
	apex.HandleFunc(func(event json.RawMessage, ctx *apex.Context) (interface{}, error) {
		envFile, err := ioutil.ReadFile(".env")
		var env Env
		json.Unmarshal(envFile, &env)
		os.Setenv("AWS_ACCESS_KEY_ID", env.AWSACCESSKEYID)
		os.Setenv("AWS_SECRET_ACCESS_KEY", env.AWSSECRETACCESSKEY)

		creds := credentials.NewEnvCredentials()

		svc := s3.New(session.New(), &aws.Config{
			Region: aws.String("us-west-2"),
			Credentials: creds})

		lsObjs := s3.ListObjectsInput{
			Bucket: aws.String("transcoderoutput489349"),
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

		return string(response), nil
	})
}
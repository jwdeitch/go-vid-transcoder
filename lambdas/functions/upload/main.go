package main

import (
	"encoding/json"
	"github.com/inturn/go-helpers"
	//"github.com/aws/aws-sdk-go/aws"
	//"github.com/aws/aws-sdk-go/service/s3"
	//"github.com/aws/aws-sdk-go/aws/session"
	//"github.com/aws/aws-sdk-go/aws/credentials"
	//"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/apex/go-apex"
	//"io/ioutil"
	//"os"
	//"fmt"
	"log"
	"os"
	"io/ioutil"
	"database/sql"
	"time"
)

type Env struct {
	AWSACCESSKEYID     string `json:"AWS_ACCESS_KEY_ID"`
	AWSSECRETACCESSKEY string `json:"AWS_SECRET_ACCESS_KEY"`
	SQLUSR             string `json:"SQL_USR"`
	SQLPASS            string `json:"SQL_PASS"`
	SQLHOST            string `json:"SQL_HOST"`
	SQLDB              string `json:"SQL_DB"`
}

type S3UploadedDocument struct {
	Records []struct {
		EventTime         time.Time `json:"eventTime"`
		EventName         string `json:"eventName"`
		RequestParameters struct {
					  SourceIPAddress string `json:"sourceIPAddress"`
				  } `json:"requestParameters"`
		S3                struct {
					  Bucket struct {
							 Name string `json:"name"`
						 } `json:"bucket"`
					  Object struct {
							 Key  string `json:"key"`
							 Size int `json:"size"`
						 } `json:"object"`
				  } `json:"s3"`
	} `json:"Records"`
}

func main() {
	apex.HandleFunc(func(event json.RawMessage, ctx *apex.Context) (interface{}, error) {
		/* logging */
		l := log.New(os.Stderr, "", 0) //write to stderr by default
		eventString, _ := event.MarshalJSON();
		l.Println(string(eventString)) // write raw event to logs

		/* env var loading */
		envFile, err := ioutil.ReadFile(".env")
		helpers.Check(err);
		var env Env
		json.Unmarshal(envFile, &env)
		os.Setenv("AWS_ACCESS_KEY_ID", env.AWSACCESSKEYID)
		os.Setenv("AWS_SECRET_ACCESS_KEY", env.AWSSECRETACCESSKEY)

		/* sql setup */
		db, err := sql.Open("mysql", env.SQLUSR + ":" + env.SQLPASS + "@" + env.SQLHOST + "/" + env.SQLDB)
		helpers.Check(err)
		defer db.Close()

		/* Upload logic (insert to RDS, and initiate ETS job */
		var s3Upload S3UploadedDocument
		json.Unmarshal(event, &s3Upload)

		insStmt, _ := db.Prepare("INSERT INTO video_service VALUES (?, ?, ?, ?, ?, ?, ?)")
		defer insStmt.Close();
		for _, s3record := range s3Upload.Records { // we can upload many vids in 1 request
			display_key := helpers.RandomString(10)
			insStmt.Exec(display_key, // p_key
				s3record.S3.Object.Key, // video title (filename)
				s3record.EventTime, // time of upload
				s3record.RequestParameters.SourceIPAddress, // uploaders IP
				nil, // length of video
				nil, // number of thumbnails generated
				true, // in processing?
				s3record.S3.Object.Size) // size of uploaded file
		}

		l.Println("completed upload")
		return event, nil

		//creds := credentials.NewEnvCredentials()
		//
		//svc := s3.New(session.New(), &aws.Config{
		//	Region: aws.String("us-west-2"),
		//	Credentials: creds})

	})
}
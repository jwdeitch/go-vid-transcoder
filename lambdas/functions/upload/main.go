package main

import (
	"encoding/json"
	"github.com/inturn/go-helpers"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/elastictranscoder"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	_ "github.com/go-sql-driver/mysql"
	"github.com/apex/go-apex"
	"log"
	"os"
	"io/ioutil"
	"database/sql"
	"time"
	"strconv"
	"strings"
)

type Env struct {
	AWSACCESSKEYID     string `json:"AWS_ACCESS_KEY_ID"`
	AWSSECRETACCESSKEY string `json:"AWS_SECRET_ACCESS_KEY"`
	SQLUSR             string `json:"SQL_USR"`
	SQLPASS            string `json:"SQL_PASS"`
	SQLHOST            string `json:"SQL_HOST"`
	SQLDB              string `json:"SQL_DB"`
	WORKINGBUCKET      string `json:"WORKING_BUCKET"`
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
		db, err := sql.Open("mysql", env.SQLUSR + ":" + env.SQLPASS + "@tcp(" + env.SQLHOST + ":3306)/" + env.SQLDB)
		if err != nil {
			l.Println("ERROR 0")
		}
		defer db.Close()

		/* S3 setup */
		os.Unsetenv("AWS_SESSION_TOKEN")
		creds := credentials.NewEnvCredentials()
		s3Service := s3.New(session.New(), &aws.Config{
			Region: aws.String("us-east-1"),
			Credentials: creds})

		ETCService := elastictranscoder.New(session.New(), &aws.Config{
			Region: aws.String("us-east-1"),
			Credentials: creds})

		/* Upload logic (insert to RDS, and initiate ETS job */
		var s3Upload S3UploadedDocument
		json.Unmarshal(eventString, &s3Upload)

		insStmt, err := db.Prepare("INSERT INTO videos VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			l.Println(err.Error())
		}

		defer insStmt.Close();
		for _, s3record := range s3Upload.Records {
			// we can upload many vids in 1 request

			currentTimeAsString := strconv.FormatInt(time.Now().Unix(), 10)
			display_key := helpers.RandomString(10)
			_, err := insStmt.Exec(display_key, // p_key
				strings.Split(s3record.S3.Object.Key, "/")[1], // video title (filename)
				s3record.EventTime, // time of upload
				s3record.RequestParameters.SourceIPAddress, // uploaders IP
				0, // length of video
				0, // number of thumbnails generated
				true, // in processing?
				s3record.S3.Object.Size, // size of uploaded file
				currentTimeAsString) // processing_timestamp
			if err != nil {
				l.Println(err);
			}

			uniqueKey := currentTimeAsString + "#" + display_key + "#";

			fileNameSlice := uniqueKey + strings.Split(s3record.S3.Object.Key, "/")[1]
			copySource := s3record.S3.Bucket.Name + "/" + s3record.S3.Object.Key

			copoutput, err := s3Service.CopyObject(&s3.CopyObjectInput{
				Bucket: aws.String(s3record.S3.Bucket.Name),
				Key: aws.String(fileNameSlice),
				CopySource: aws.String(copySource)})
			if err != nil {
				l.Println(err.Error())
			} else {
				l.Println(copoutput.String())
			}

			deloutput, err := s3Service.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(s3record.S3.Bucket.Name),
				Key: aws.String(s3record.S3.Object.Key)})

			if err != nil {
				l.Println(err.Error())
			} else {
				l.Println(deloutput.String())
			}

			transcodedOutputKey := "output/" + uniqueKey + ".webm"
			thumnbPattern := "output/" + uniqueKey + "_thumb{count}"
			etcResponse, err := ETCService.CreateJob(&elastictranscoder.CreateJobInput{
				Input: &elastictranscoder.JobInput{
					Key: aws.String(fileNameSlice)},
				PipelineId: aws.String("1469293642428-kiypmq"),
				Output: &elastictranscoder.CreateJobOutput{
					PresetId:aws.String("1469295414594-fgaog2"),
					Key:aws.String(transcodedOutputKey),
					ThumbnailPattern: aws.String(thumnbPattern)}})

			if err != nil {
				l.Println(err.Error())
			} else {
				l.Println(etcResponse.String())
			}

		}

		l.Println("completed upload")
		return event, nil

	})
}
package main

import (
	"encoding/json"
	"github.com/inturn/go-helpers"
	_ "github.com/go-sql-driver/mysql"
	"github.com/apex/go-apex"
	"log"
	"os"
	"io/ioutil"
	"strings"
	"math"
	"time"
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

type conditions struct {
	Bucket string `json:"bucket,omitempty"`
	Acl string `json:"acl,omitempty"`
	SuccessActionRedirect string `json:"success_action_redirect,omitempty"`
	XAmzCredential string `json:"x-amz-credential,omitempty"`
	XAmzAlgorithm string `json:"x-amz-algorithm,omitempty"`
	XAmzDate string `json:"x-amz-date,omitempty"`
}

type policy struct {
	Expiration time.Time `json:"expiration"`
	Conditions []conditions `json:"conditions"`
}

type clientParams struct {
	Filename string `json:"file"`
	Type string `json:"type"`
	Size string `json:"size"`
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

		var ClientPostParams clientParams
		json.Unmarshal(eventString, &ClientPostParams)

		currentUnixTime := time.Now().Unix() + 10800 // 3 hrs

		clientPolicy := policy{
			Expiration: time.Unix(currentUnixTime, 0),
			Conditions: conditions{
				Acl: "public-read",
				SuccessActionRedirect:"http://tv.rsa.pub/",
			XAmzCredential:},

		}

		l.Println("Recieved Transcode Job")
		return event, nil

	})
}
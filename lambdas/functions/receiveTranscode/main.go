package main

import (
	"encoding/json"
	"github.com/inturn/go-helpers"
	_ "github.com/go-sql-driver/mysql"
	"github.com/apex/go-apex"
	"log"
	"os"
	"io/ioutil"
	"database/sql"
	"strings"
	"math"
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

type SNSOutput struct {
	Records []struct {
		Sns struct {
			    Message string `json:"Message"`
		    } `json:"Sns"`
	} `json:"Records"`
}

type ETSMessage struct {
	JobID   string `json:"jobId"`
	State   string `json:"state"`
	Input   struct {
			Key string `json:"key"`
		} `json:"input"`
	Outputs []struct {
		Status   string `json:"status"`
		Duration int `json:"duration"`
		Width    int `json:"width"`
		Height   int `json:"height"`
	} `json:"outputs"`
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

		/* sql setup */
		db, err := sql.Open("mysql", env.SQLUSR + ":" + env.SQLPASS + "@tcp(" + env.SQLHOST + ":3306)/" + env.SQLDB)
		if err != nil {
			l.Println("ERROR 0")
		}
		defer db.Close()

		var snsOutput SNSOutput
		json.Unmarshal(eventString, &snsOutput)

		var etsMessage ETSMessage
		message := snsOutput.Records[0].Sns.Message
		message = strings.Replace(message, "\n", "", -1)
		message = strings.Replace(message, "\\", "", -1)
		json.Unmarshal([]byte(message), &etsMessage)

		p_key := strings.Split(etsMessage.Input.Key, "#")[1]

		if etsMessage.State != "COMPLETED" {
			db.Prepare("DELETE FROM videos WHERE display_key = ?")
			db.Exec(p_key)
			return event, nil
		}

		uptStmt, err := db.Prepare("UPDATE videos SET processing = false, length = ?, thumb_count = ? WHERE display_key = ?")
		if err != nil {
			l.Println(err.Error())
		}
		defer uptStmt.Close();

		_, error := uptStmt.Exec(etsMessage.Outputs[0].Duration,
			math.Ceil(float64((etsMessage.Outputs[0].Duration / 10) + 1)),
			p_key)

		if error != nil {
			l.Println(error.Error())
		}

		l.Println("Recieved Transcode Job")
		return event, nil

	})
}
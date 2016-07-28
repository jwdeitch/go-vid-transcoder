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

type DbRow struct {
	DisplayKey       string
	Name             string
	Uploaded_at      string
	Length           int64
	ThumbCount       int32
	Processing       bool
	PreTranscodeSize int64
	Stamp            int64
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

		rows, err := db.Query("SELECT * FROM videos order by processing desc, uploaded_at desc LIMIT 200")
		if err != nil {
			l.Println(err.Error())
		}
		var Rows []DbRow
		for rows.Next() {
			var d_key string
			var name string
			var uploaded_at string
			var uploaded_by string
			var length int64
			var thumb_count int32
			var processing bool
			var size int64
			var timestamp int64
			err = rows.Scan(&d_key, &name, &uploaded_at, &uploaded_by, &length, &thumb_count, &processing, &size, &timestamp)

			Rows = append(Rows, DbRow{d_key, name, uploaded_at, length, thumb_count, processing, size, timestamp})
		}

		return Rows, nil

	})
}
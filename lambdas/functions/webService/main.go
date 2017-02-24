package main

import (
	"encoding/json"
	"github.com/inturn/go-helpers"
	"github.com/apex/go-apex"
	"os"
	"io/ioutil"
	"regexp"
	"strconv"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

type Env struct {
	AWSACCESSKEYID     string `json:"AWS_ACCESS_KEY_ID"`
	AWSSECRETACCESSKEY string `json:"AWS_SECRET_ACCESS_KEY"`
	SQLUSR             string `json:"SQL_USR"`
	SQLPASS            string `json:"SQL_PASS"`
	SQLHOST            string `json:"SQL_HOST"`
	SQLDB              string `json:"SQL_DB"`
	SQLPORT            string `json:"SQL_PORT"`
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
	Notes            string
}

type StatRow struct {
	T_length int64
	T_size   int64
	T_count  int
	T_users  int
}

type query struct {
	Params struct {
		       Querystring struct {
					   Q     *string `json:"q,omitempty"`
					   Limit string `json:"limit"`
					   Skip  string `json:"skip"`
				   } `json:"querystring"`
	       } `json:"params"`
}

func main() {
	apex.HandleFunc(func(event json.RawMessage, ctx *apex.Context) (interface{}, error) {
		/* logging */
		l := log.New(os.Stderr, "", 0) //write to stderr by default
		eventString, _ := event.MarshalJSON();
		l.Println(string(eventString)) // write raw event to logs

		/* env var loading */
		envFile, err := ioutil.ReadFile(".env")
		helpers.ErrorCheckAndPrint(err);
		var env Env
		json.Unmarshal(envFile, &env)
l.Println("user=" + env.SQLUSR +
	" dbname=" + env.SQLDB +
	" dbhost=" + env.SQLHOST +
	" dbpass=" + env.SQLPASS +
	" dbport=" + env.SQLPORT +
	" sslmode=enable");
		/* sql setup */
		db, err := sql.Open("postgres",
			"user=" + env.SQLUSR +
				" dbname=" + env.SQLDB +
				" dbhost=" + env.SQLHOST +
				" dbpass=" + env.SQLPASS +
				" dbport=" + env.SQLPORT +
				" sslmode=enable")
		if err != nil {
			l.Println("SQL ERROR 0", err.Error())
		}
		defer db.Close()

		var rows *sql.Rows
		var statRows *sql.Rows

		var searchQuery query;
		json.Unmarshal(eventString, &searchQuery)

		reg, err := regexp.Compile("[^A-Za-z0-9]+")
		if err != nil {
			log.Fatal(err)
		}

		skip, err := strconv.Atoi(searchQuery.Params.Querystring.Skip)
		limit, err := strconv.Atoi(searchQuery.Params.Querystring.Limit)

		if (searchQuery.Params.Querystring.Q != nil) {
			keywords := "%" + reg.ReplaceAllString(*searchQuery.Params.Querystring.Q, "_") + "%"
			rows, err = db.Query("SELECT * FROM videos where video_title ilike ? and is_private = false order by is_processing desc, uploaded_at desc LIMIT ?, ?", keywords, skip, limit)

			statRows, err = db.Query("select sum(video_length_seconds) as t_length, sum(pre_transcode_size_bytes) as t_size, count(*) as t_count, count(distinct uploaded_by) as t_users FROM videos where video_title ilike ? and is_private = false", keywords, skip, limit)

		} else {
			rows, err = db.Query("SELECT * FROM videos where is_private = false order by processing desc, uploaded_at desc LIMIT ?, ?", skip, limit)

			statRows, err = db.Query("select sum(video_length_seconds) as t_length, sum(pre_transcode_size_bytes) as t_size, count(*) as t_count, count(distinct uploaded_by) as t_users FROM videos where is_private = false")
		}

		if err != nil {
			l.Println(err.Error())
		}
		var Rows []interface{}
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
			var notes string
			var private string
			err = rows.Scan(&d_key, &name, &uploaded_at, &uploaded_by, &length, &thumb_count, &processing, &size, &timestamp, &notes, &private)

			Rows = append(Rows, DbRow{d_key, name, uploaded_at, length, thumb_count, processing, size, timestamp, notes})
		}

		for statRows.Next() {
			var t_length int64
			var t_size int64
			var t_count int
			var t_users int
			err = statRows.Scan(&t_length, &t_size, &t_count, &t_users)
			Rows = append(Rows, StatRow{t_length, t_size, t_count, t_users})
		}

		return Rows, nil

	})
}
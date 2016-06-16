package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	setupDB()
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", welcome).Methods("GET")
	router.HandleFunc("/data", getdata).Methods("GET")
	router.HandleFunc("/data", postdata).Methods("POST")
	port := ":8888"
	log.Printf("Listening at %s", port)
	log.Fatal(http.ListenAndServe(port, router))
}

type Likes struct {
	Count int
}

type Response struct {
	Success bool
	Message string
	Data    *Likes
}

func welcome(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Welcome to postgres sample, Datapoints available: \n /data[GET] \n /data[POST]")
}

func getdata(w http.ResponseWriter, r *http.Request) {
	log.Println(r.UserAgent())
	resp := new(Response)
	db, err := dbConnection()
	if err != nil {
		log.Fatal(err)
		sendErr(err, w)
		return
	}
	like := new(Likes)
	err = db.QueryRow("select * from tbl_Counter").Scan(&like.Count)
	resp.Data = like

	sendResp(resp, err, w)
}

func postdata(w http.ResponseWriter, r *http.Request) {
	log.Println(r.UserAgent())
	var (
		body []byte
		err  error
		like Likes
		db   *sql.DB
	)
	if r.Body != nil {
		if body, err = ioutil.ReadAll(r.Body); err == nil {
			log.Println("----- dd ", string(body))
			if err = json.Unmarshal(body, &like); err == nil {
				log.Println("----- Input ", like)
				if like.Count >= 0 {
					db, err = dbConnection()
					if err == nil {
						_, err = db.Query(fmt.Sprintf("update tbl_Counter set likecount = %d;", like.Count))
					}
				} else {
					err = fmt.Errorf("Like should be greater than 0.")
				}
			}
		}
	}
	resp := new(Response)
	sendResp(resp, err, w)

}
func sendErr(err error, w http.ResponseWriter) {
	resp := new(Response)
	sendResp(resp, err, w)
}
func sendResp(resp *Response, err error, w http.ResponseWriter) {
	resp.Success = true
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var out []byte
	if err != nil {
		resp.Success = false
		resp.Message = err.Error()
		out, _ = json.MarshalIndent(resp, "", " ")
		fmt.Fprintln(w, string(out))
		return
	}
	out, err = json.MarshalIndent(resp, "", " ")
	//In Any case print error
	if err != nil {
		resp.Success = false
		resp.Message = err.Error()
		fmt.Fprintln(w, resp)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, string(out))
}
func dbConnection() (db *sql.DB, err error) {
	db, err = sql.Open("postgres", "postgres://postgres:postgres@test--pgtest--pgsingle--1164ae-0.service.consul:4000/postgresDB?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	return
}

func setupDB() error {
	// Set DB
	db, err := dbConnection()
	if err == nil {
		_, err = db.Query("CREATE TABLE IF NOT EXISTS tbl_Counter (likecount bigint)")
		if err == nil {
			_, err = db.Query("insert into tbl_Counter values (0)")
		}

	}
	if err != nil {
		log.Fatal(err)
	}
	return err
}

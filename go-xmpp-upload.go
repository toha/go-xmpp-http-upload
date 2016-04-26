// Copyright 2016 Tobias Hartwich (tobias.hartwich@gmail.com).
// All rights reserved. Use of this source code is governed by the
// MIT license that can be found in the LICENSE file.

package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var allowedSlotIp = make(map[string]int)
var db *sql.DB = nil // database connection pool
var basePutURL string
var baseGetURL string
var uploadDir string

type Upload struct {
	id            int
	slot_hash     string
	jid           string
	original_name string
	disk_name     string
	upload_time   *time.Time
	file_size     int
	content_type  string
	slot_time     *time.Time
}

func registerSlotHandler(w http.ResponseWriter, r *http.Request) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Println("Nicht erlaubt")
		http.Error(w, "Not Allowed", http.StatusForbidden)
		return
	}

	i, ok := allowedSlotIp[ip]
	if !ok || i != 1 {
		fmt.Println("Nicht erlaubt")
		http.Error(w, "Not Allowed", http.StatusForbidden)
		return
	}

	r.ParseForm()

	var jid string = r.FormValue("jid")
	var fileName string = strings.Replace(r.FormValue("name"), " ", "_", -1)
	var contentType string = r.FormValue("content_type")
	fileSize, err := strconv.Atoi(r.FormValue("size"))
	if err != nil {
		fmt.Println("Invalid Size")
		http.Error(w, "Invalid Size", http.StatusInternalServerError)
		return
	}

	if len(jid) == 0 || len(fileName) == 0 || len(contentType) == 0 || fileSize == 0 {
		fmt.Println("Invalid Parameter")
		http.Error(w, "Invalid Parameter", http.StatusBadRequest)
		return
	}

	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var randomHash string = fmt.Sprintf("%x", sha256.Sum256(randomBytes))
	var diskName string = randomHash + "_" + fileName

	stmt, err := db.Prepare("INSERT INTO uploads (slot_hash, jid, original_name, disk_name, upload_time, file_size, content_type, slot_time) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = stmt.Exec(randomHash, jid, fileName, diskName, nil, fileSize, contentType, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, basePutURL+"\n", randomHash, fileName)
	fmt.Fprintf(w, baseGetURL+"\n", randomHash, fileName)

	fmt.Printf(basePutURL+"\n", randomHash, fileName)
	fmt.Printf(baseGetURL+"\n", randomHash, fileName)

}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method != "PUT" {
		http.Error(w, "Invalid Parameter", http.StatusBadRequest)
		return
	}

	fmt.Println(r.URL.Path)
	var validPath = regexp.MustCompile("^/upload/([a-f0-9]+)/(.*?)$")
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.Error(w, "Invalid Path", http.StatusBadRequest)
		return
	}

	// Get Upload from DB
	bk := new(Upload)
	rows, err := db.Query("SELECT * FROM uploads WHERE slot_hash = $1 AND upload_time IS NULL", m[1])
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows.Next()
	errScan := rows.Scan(&bk.id, &bk.slot_hash, &bk.jid, &bk.original_name, &bk.disk_name, &bk.upload_time, &bk.file_size, &bk.content_type, &bk.slot_time)
	if errScan != nil {
		log.Fatal(errScan)
	}

	f, err := os.OpenFile(path.Join(uploadDir, bk.disk_name), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, r.Body)

	// Update upload-DAte
	stmt, err := db.Prepare("UPDATE uploads SET upload_time = $1 WHERE id = $2")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = stmt.Exec(time.Now(), bk.id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	var validPath = regexp.MustCompile("^/download/([a-f0-9]+)/(.*?)$")
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.Error(w, "Invalid Path", http.StatusBadRequest)
		return
	}

	// Get Upload from DB
	bk := new(Upload)
	rows, err := db.Query("SELECT * FROM uploads WHERE slot_hash = $1 AND upload_time IS NOT NULL", m[1])
	if err != nil {
		http.Error(w, "Invalid Hash", http.StatusBadRequest)
		return
	}
	defer rows.Close()
	rows.Next()
	errScan := rows.Scan(&bk.id, &bk.slot_hash, &bk.jid, &bk.original_name, &bk.disk_name, &bk.upload_time, &bk.file_size, &bk.content_type, &bk.slot_time)
	if errScan != nil {
		http.Error(w, "Invalid Hash", http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, path.Join(uploadDir, bk.disk_name))

}

func main() {

	listeningString := os.Getenv("XMPP_UPLOAD_LISTEN")
	if len(listeningString) == 0 {
		listeningString = ":8080"
	}

	ejabberdAddress := os.Getenv("EJABBERD_PORT_5222_TCP_ADDR")
	if len(ejabberdAddress) == 0 {
		log.Fatal("env var EJABBERD_PORT_5222_TCP_ADDR missing")
		return
	}
	allowedSlotIp[ejabberdAddress] = 1

	postgresAddress := os.Getenv("POSTGRES_PORT_5432_TCP_ADDR")
	if len(postgresAddress) == 0 {
		log.Fatal("env var POSTGRES_PORT_5432_TCP_ADDR missing")
		return
	}

	postgresUser := os.Getenv("POSTGRES_USER")
	if len(postgresUser) == 0 {
		log.Fatal("env var POSTGRES_USER missing")
		return
	}

	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	if len(postgresPassword) == 0 {
		log.Fatal("env var POSTGRES_PASSWORD missing")
		return
	}

	postgresDatabase := os.Getenv("POSTGRES_DATABASE")
	if len(postgresDatabase) == 0 {
		log.Fatal("env var POSTGRES_DATABASE missing")
		return
	}

	uploadDir = os.Getenv("UPLOADED_FILES_DIR")
	if len(uploadDir) == 0 {
		uploadDir = "/opt/xmpp_uploads"
	}
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		log.Fatal(fmt.Sprintf("Upload directory '%s' does not exist", uploadDir))
		return
	}

	putGetUrlHost := os.Getenv("PUT_GET_URL_HOST")
	if len(putGetUrlHost) == 0 {
		log.Fatal("env var PUT_GET_URL_HOST missing")
		return
	}
	basePutURL = putGetUrlHost + "/upload/%s/%s"
	baseGetURL = putGetUrlHost + "/download/%s/%s"

	envAllowedIPs := os.Getenv("ALLOWED_IPS")
	if len(envAllowedIPs) != 0 {
		ipsSplitted := strings.Split(envAllowedIPs, ",")
		for _, ip := range ipsSplitted {
			allowedSlotIp[ip] = 1
		}
	}

	postgresConnectionString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", postgresAddress, postgresUser, postgresPassword, postgresDatabase)

	var err error
	db, err = sql.Open("postgres", postgresConnectionString)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	/*rows, err := db.Query("SELECT * FROM uploads")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	bks := make([]*Upload, 0)
	for rows.Next() {
		bk := new(Upload)
		err := rows.Scan(&bk.id, &bk.slot_hash, &bk.jid, &bk.original_name, &bk.disk_name, &bk.upload_time, &bk.file_size, &bk.content_type, &bk.slot_time)
		if err != nil {
			log.Fatal(err)
		}
		bks = append(bks, bk)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	for _, bk := range bks {
		fmt.Printf("%d, %s, %s\n", bk.id, bk.slot_hash, bk.original_name)
	}*/

	http.HandleFunc("/slot", registerSlotHandler)
	http.HandleFunc("/upload/", uploadHandler)
	http.HandleFunc("/download/", downloadHandler)
	//http.Handle("/", http.FileServer(http.Dir("./public")))

	err = http.ListenAndServe(listeningString, nil)
	fmt.Println(err)
}

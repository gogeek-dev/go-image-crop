package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

func dbConns() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "welcome123$"
	dbName := "goblog"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

type images struct {
	FileName, Path string
	ID             int
}

func savefile(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {

		imgBase := r.FormValue("data")

		// remove "data:image/png;base64,"
		imgBase64cleaned := imgBase[len("data:image/png;base64,"):len(imgBase)]

		// decode base64 to buffer bytes
		imgBytes, _ := base64.StdEncoding.DecodeString(imgBase64cleaned)

		tempFile, err := ioutil.TempFile("cropped_image", "cropimg-*.png")
		if err != nil {
			fmt.Println(err)
		}
		db := dbConns()
		path := tempFile.Name() //file storing path
		dir, FileName := filepath.Split(path)

		fmt.Println("Dir:", dir)
		fmt.Println("File Name:", FileName)
		insert, err := db.Prepare("INSERT INTO imgcrop(filename,path)VALUES(?,?)")
		if err != nil {

			fmt.Printf("not inserted")
		}
		insert.Exec(FileName, path)

		defer db.Close()
		defer tempFile.Close()

		tempFile.Write(imgBytes)

	}
}

func cropimage(w http.ResponseWriter, r *http.Request) {

	t, _ := template.ParseFiles("cropimage.html")
	t.Execute(w, nil)
}

func imagelist(w http.ResponseWriter, r *http.Request) {
	db := dbConns()
	selDB, err := db.Query("SELECT * FROM imgcrop ORDER BY id DESC")
	if err != nil {
		panic(err.Error())
	}
	img := images{}
	res := []images{}
	for selDB.Next() {
		var id int
		var filename, path string
		err = selDB.Scan(&id, &filename, &path)
		if err != nil {
			panic(err.Error())
		}
		img.ID = id
		img.FileName = filename

		img.Path = path

		res = append(res, img)
	}
	t, _ := template.ParseFiles("imglist.html")
	t.Execute(w, res)
	defer db.Close()
}

func main() {

	http.HandleFunc("/", cropimage)
	http.HandleFunc("/savecroppedphoto", savefile)
	http.HandleFunc("/imagelist", imagelist)
	http.Handle("/cropped_image/", http.StripPrefix("/cropped_image/", http.FileServer(http.Dir("cropped_image"))))

	http.Handle("/file/", http.StripPrefix("/file/", http.FileServer(http.Dir("file"))))

	fmt.Println("listening on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

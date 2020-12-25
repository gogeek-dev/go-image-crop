package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func dbConns() (db *sql.DB) {
	er := godotenv.Load(".env")

	if er != nil {
		log.Fatalf("Error loading .env file")
	}
	dbDriver := os.Getenv("DB_DRIVER")
	dbUser := os.Getenv("DB_USERNAME")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
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

func addcrop(w http.ResponseWriter, r *http.Request) {

	t, _ := template.ParseFiles("templates/cropimage.html")
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
	t, _ := template.ParseFiles("templates/imglist.html")
	t.Execute(w, res)
	defer db.Close()
}

func main() {

	http.HandleFunc("/", imagelist)
	http.HandleFunc("/savecroppedphoto", savefile)
	http.HandleFunc("/addcrop", addcrop)
	http.Handle("/cropped_image/", http.StripPrefix("/cropped_image/", http.FileServer(http.Dir("cropped_image"))))

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	fmt.Println("listening on http://localhost:8090")
	http.ListenAndServe(":8090", nil)
}

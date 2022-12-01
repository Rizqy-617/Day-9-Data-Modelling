package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"routing/connection"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)


func main() {
	route := mux.NewRouter()

	connection.DatabaseConnect()

	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public/"))))

	route.HandleFunc("/", home).Methods("GET")
	route.HandleFunc("/project", project).Methods("GET")
	route.HandleFunc("/contact", contact).Methods("GET")
	route.HandleFunc("/addProject", addProject).Methods("POST")
	route.HandleFunc("/projectDetail/{id}", projectDetail).Methods("GET")
	route.HandleFunc("/editProject/{id}", editProject).Methods("GET")
	route.HandleFunc("/updateProject/{index}", updateProject).Methods(("POST"))
	route.HandleFunc("/deleteProject/{index}", deleteProject).Methods("GET")

	fmt.Println("Server running on port 5000")
	http.ListenAndServe("localhost:5000", route)
}

//Buat Home
func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("view/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	// dataProject := map[string]interface{}{
	// 	"Projects": projectSubmit,
	// }

	dataProject, errQuery := connection.Conn.Query(context.Background(), "SELECT id, project_name, description, technologies, image FROM tb_project")
	if errQuery!= nil {
		fmt.Println("Message : " + errQuery.Error())
		return
	}

	var result []Projectsubmit

	for dataProject.Next() {
		var each = Projectsubmit{}

		err := dataProject.Scan(&each.ID, &each.Projectname, &each.Description, &each.Technologies, &each.Image)
		if err != nil {
			fmt.Println("Message : " + err.Error())
			return
		}

		result = append(result, each)
	}

	myProject := map[string]interface{} {
		"Projects": result,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, myProject)
}

//Buat Project
func project(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("view/project.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	tmpl.Execute(w, nil)
}

// struct submit
type Projectsubmit struct {
	ID int
	Projectname string
	Description string
	Technologies []string
	Startdate string
	Enddate string
	Image string
	Duration string
}

// var buat projectsubmit nya
var projectSubmit = []Projectsubmit{
	/* Ini cuman buat nampilin data secara statis/data dummy
	{
	Projectname: "Nama Project",
	Description: "Ini Deskripsi",
	},
	*/
}


//Buat ngeappend data project ke home
func addProject(w http.ResponseWriter, r*http.Request) {
	err := r.ParseMultipartForm(1024)
	if err != nil {
		log.Fatal(err)
	}

	projectName := r.PostForm.Get("project_name")
	description := r.PostForm.Get("description")
	technologies := r.Form["technologies"]
	startDate := r.PostForm.Get("startDate")
	endDate := r.PostForm.Get("endDate")

	//Buat Durasi
	const timeFormat = "2006-01-02"
	timeStartdate, err := time.Parse(timeFormat, startDate)//Mengformat startDate menjadi format yang ada di timeFormat
	if err !=  nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeEnddate, err := time.Parse(timeFormat, endDate)//Mengformat endDate menjadi format yang ada di timeFormat
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	selisihHari := timeEnddate.Sub(timeStartdate)// Untuk menghitung selisih hari
	hari := int64(selisihHari.Hours() / 24) //Untuk mengubah format "N"h"N"m"N"s menjadi format jam dan dibagi 24 jam dan mengset menjadi tipedata int64

	var duration string
	if hari >= 0 {
		duration = strconv.FormatInt(hari, 10) + " hari" //Mengconvert int64 menjadi string
	}

	//Durasi end


	//Image Start
	img, imgname, err := r.FormFile("image") //Buat ngambil data dari form
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer img.Close()
	dir, err := os.Getwd()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filename := imgname.Filename //buat ngebuat nama file nya
	fileLocation := filepath.Join(dir, "public/uploadedImage", filename) //buat masukin file ke dalam folder
	targetFile, err := os.OpenFile(fileLocation, os.O_WRONLY|os.O_CREATE, 0666)//memindahkan file yang sudah di copy ke dalam fileLocation
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer targetFile.Close()
	if _, err := io.Copy(targetFile, img); err != nil { //mengcopy file img ke targetFile
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//Image CLose

	//Buat Ngeupload file
	var newProject = Projectsubmit{
		Projectname: projectName,
		Description: description,
		Technologies: technologies,
		Startdate: startDate,
		Enddate: endDate,
		Image: imgname.Filename, // didalam image harus diisi handler.Filename biar yang dipanggilnya filename nya
		Duration: duration,
	}

	/* Buat nampilin ke terminal
	fmt.Println("Ini project name : " + r.PostForm.Get("project_name"))
	fmt.Println("Ini start-date : " + r.PostForm.Get("startDate"))
	fmt.Println("Ini end-date : " + r.PostForm.Get("endDate"))
	fmt.Println("Ini deskripsi : " + r.PostForm.Get("description"))
	fmt.Println("Teknologi yang digunakan : ", r.Form["technologies"])
	fmt.Println(r.PostForm.Get("image"))
	*/

	//Buat ngeappend newProject ke dalam variable projectSubmit
	projectSubmit = append(projectSubmit, newProject)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

//Fungsi buat nampilin data project di halamannya sendiri
func projectDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpl, err := template.ParseFiles("view/project-detail.html")

	if err != nil {
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	id, _ :=strconv.Atoi(mux.Vars(r)["id"])

	ProjectDetail := Projectsubmit{}

	for index, data := range projectSubmit {
		if index == id {
			ProjectDetail = Projectsubmit{
				ID: id,
				Projectname: data.Projectname,
				Description: data.Description,
				Technologies: data.Technologies,
				Startdate: data.Startdate,
				Enddate: data.Enddate,
				Duration: data.Duration,
				Image: data.Image,
		    }
		}
	}

	fmt.Println(ProjectDetail)

	dataDetail := map[string]interface{}{
		"Projectsubmit": ProjectDetail,
	}

	tmpl.Execute(w, dataDetail)
}

//editProject
func editProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpl, err := template.ParseFiles("view/edit-project.html")

	if err != nil {
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	id, _ :=strconv.Atoi(mux.Vars(r)["id"])

	ProjectDetail := Projectsubmit{}

	for index, data := range projectSubmit {
		if index == id {
			ProjectDetail = Projectsubmit{
				ID: id,
				Projectname: data.Projectname,
				Description: data.Description,
				Technologies: data.Technologies,
				Startdate: data.Startdate,
				Enddate: data.Enddate,
				Duration: data.Duration,
				Image: data.Image,
		    }
		}
	}

	fmt.Println(ProjectDetail)

	dataDetail := map[string]interface{}{
		"Projectsubmit": ProjectDetail,
	}

	tmpl.Execute(w, dataDetail)
}

//updateProject
func updateProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(1024)
	if err != nil {
		log.Fatal(err)
	}

	projectName := r.PostForm.Get("project_name")
	description := r.PostForm.Get("description")
	technologies := r.Form["technologies"]
	startDate := r.PostForm.Get("startDate")
	endDate := r.PostForm.Get("endDate")

	//Buat Durasi
	const timeFormat = "2006-01-02"
	timeStartdate, err := time.Parse(timeFormat, startDate)//Mengformat startDate menjadi format yang ada di timeFormat
	if err !=  nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeEnddate, err := time.Parse(timeFormat, endDate)//Mengformat endDate menjadi format yang ada di timeFormat
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	selisihHari := timeEnddate.Sub(timeStartdate)// Untuk menghitung selisih hari
	hari := int64(selisihHari.Hours() / 24) //Untuk mengubah format "N"h"N"m"N"s menjadi format jam dan dibagi 24 jam dan mengset menjadi tipedata int64

	var duration string
	if hari >= 0 {
		duration = strconv.FormatInt(hari, 8) + " hari" //Mengconvert int64 menjadi string
	}

	//Durasi end


	//Image Start
	img, imgname, err := r.FormFile("image") //Buat ngambil data dari form
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer img.Close()
	dir, err := os.Getwd()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filename := imgname.Filename //buat ngebuat nama file nya
	fileLocation := filepath.Join(dir, "public/uploadedImage", filename) //buat masukin file ke dalam folder
	targetFile, err := os.OpenFile(fileLocation, os.O_WRONLY|os.O_CREATE, 0666)//memindahkan file yang sudah di copy ke dalam fileLocation
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer targetFile.Close()
	if _, err := io.Copy(targetFile, img); err != nil { //mengcopy file img ke targetFile
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	index, _ := strconv.Atoi(mux.Vars(r)["index"])
	projectSubmit[index].Projectname = projectName
	projectSubmit[index].Description = description
	projectSubmit[index].Duration = duration
	projectSubmit[index].Startdate = startDate
	projectSubmit[index].Enddate = endDate
	projectSubmit[index].Technologies = technologies
	projectSubmit[index].Image = imgname.Filename

	http.Redirect(w, r, "/", http.StatusMovedPermanently )
}

//Fungsi buat ngedelete
func deleteProject(w http.ResponseWriter, r *http.Request) {
	index, _ := strconv.Atoi(mux.Vars(r)["index"])

	projectSubmit = append(projectSubmit[:index], projectSubmit[index+1:]...)

	http.Redirect(w, r, "/", http.StatusFound)
}

//Buat Contact
func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("view/contact.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)
}




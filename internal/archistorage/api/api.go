package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	model "gitlab.archimed-soft.ru/web/archistorage/internal/archistorage/models"
	"gitlab.archimed-soft.ru/web/archistorage/internal/archistorage/utils"
)

const deltaTimeSec = 30
const serverPort = "2332"

var pathSeparator string

type ServerHandler struct {
	UrlFiles        map[string]model.FileInfo
	LinkFiles       map[string]string
	MedcartPhotoDir *string
	InfoLog         *log.Logger
	ErrorLog        *log.Logger
	ServerPort      string
	Config          *model.IniConfig
	WebServer       *http.Server
	IsDebugMode     bool
}

func (s *ServerHandler) getPathFormUUID(uuID string) string {
	return uuID[0:2] + string(os.PathSeparator) + uuID[3:5]
}

func (s *ServerHandler) runWebServer() {
	router := mux.NewRouter()
	router.HandleFunc("/v1/storage", s.info).Methods("GET")
	router.HandleFunc("/v1/storage/upload", s.upload).Methods("POST")
	router.HandleFunc("/v1/storage/upload-target/{key}", s.targetUpload).Methods("PUT")
	router.HandleFunc("/v1/storage/download/{key}", s.download).Methods("GET")
	router.HandleFunc("/v1/storage/download-target/{key}", s.targetDownload).Methods("GET")
	router.HandleFunc("/v1/storage/delete/{key}", s.delete).Methods("DELETE")

	s.WebServer = &http.Server{Addr: ":" + s.ServerPort, Handler: router}
	s.InfoLog.Printf("Запуск сервера на %s", "localhost:"+s.ServerPort)
	fmt.Printf("Запуск сервера на %s\n", "localhost:"+s.ServerPort)
	fmt.Printf("http://localhost:2332/v1/storage - info\n")
	err := s.WebServer.ListenAndServe()

	s.ErrorLog.Fatal(err)
}

func (s *ServerHandler) Run(inThread bool) {

	pathSeparator = string(os.PathSeparator)

	s.ServerPort = serverPort

	if inThread {
		go s.runWebServer()
	} else {
		s.runWebServer()
	}
}

func (s *ServerHandler) info(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	answer := make(map[string]string)
	answer["appName"] = "Archimed+ storage files"
	answer["version"] = "1.0"
	err := json.NewEncoder(w).Encode(answer)
	if err != nil {
		s.ErrorLog.Fatal(err)
	}
}

func (s *ServerHandler) upload(w http.ResponseWriter, r *http.Request) {

	var info model.FileInfo
	json.NewDecoder(r.Body).Decode(&info)

	addrS := strings.Split(r.RemoteAddr, ":")

	if len(addrS) == 0 {
		err := errors.New("")
		s.ErrorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error()))
		return
	}

	uuidWithHyphen := uuid.New()
	uuID := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
	path := s.Config.MedcartPhotoDir + pathSeparator + s.getPathFormUUID(uuID)
	err := os.MkdirAll(path, 0777)
	if err != nil {
		s.ErrorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error()))
		return
	}

	info.CreatedAt = time.Now().Unix()
	fileInfo := path + pathSeparator + uuID + ".info"
	data, _ := json.Marshal(info)
	file, err := os.Create(fileInfo)
	if err != nil {
		s.ErrorLog.Println(err)
		return
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		s.ErrorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error()))
		return
	}

	info.Url = "http://" + r.Host + "/v1/storage/upload-target/" + uuID

	answer := make(map[string]string)
	answer["UUID"] = uuID
	answer["href"] = info.Url
	answer["method"] = "PUT"

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answer)
}

func (s *ServerHandler) targetUpload(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	uuID := params["key"]
	if len(uuID) < 32 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Не задан UUID файла!"))
		return
	}

	path := s.Config.MedcartPhotoDir + pathSeparator + s.getPathFormUUID(uuID)
	pathBin := path + pathSeparator + uuID + ".bin"
	pathInfo := path + pathSeparator + uuID + ".info"
	if !utils.FileExists(pathInfo) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("502 - URL недействителен!"))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.ErrorLog.Println(err)
		return
	}

	file, err := os.Create(pathBin)
	if err != nil {
		s.ErrorLog.Println(err)
		return
	}
	defer file.Close()
	_, err = file.Write(body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error()))
		return
	}
}

func (s *ServerHandler) download(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	uuID := params["key"]
	if uuID == "" {
		s.ErrorLog.Println("Не задан идентификатор файла!")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Не задан идентификатор файла!"))
		return
	}

	path := s.Config.MedcartPhotoDir + pathSeparator + s.getPathFormUUID(uuID)
	pathInfo := path + pathSeparator + uuID + ".info"

	if !utils.FileExists(pathInfo) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("502 - URL недействителен! Информация о файле не найдена!"))
		return
	}

	data, err := os.ReadFile(pathInfo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error()))
		return
	}

	var info model.FileInfo

	err = json.Unmarshal(data, &info)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - " + err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (s *ServerHandler) targetDownload(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	uuID := params["key"]
	if uuID == "" {
		s.ErrorLog.Println("Не задан идентификатор файла!")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Не задан идентификатор файла!"))
		return
	}

	path := s.Config.MedcartPhotoDir + pathSeparator + s.getPathFormUUID(uuID)
	pathBin := path + pathSeparator + uuID + ".bin"
	if !utils.FileExists(pathBin) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("502 - URL недействителен! Файл не найден!"))
		return
	}

	data, err := ioutil.ReadFile(pathBin)
	if err != nil {
		s.ErrorLog.Println(err)
	}
	w.Write(data)
}

func (s *ServerHandler) delete(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	uuID := params["key"]
	if uuID == "" {
		s.ErrorLog.Println("Не задан идентификатор файла!")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Не задан идентификатор файла!"))
		return
	}

	path := s.Config.MedcartPhotoDir + pathSeparator + s.getPathFormUUID(uuID)
	pathBin := path + pathSeparator + uuID + ".bin"
	pathInfo := path + pathSeparator + uuID + ".info"

	if utils.FileExists(pathInfo) {
		err := os.Remove(pathInfo)
		if err != nil {
			s.ErrorLog.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - " + err.Error()))
			return
		}
	}

	if utils.FileExists(pathBin) {
		err := os.Remove(pathBin)
		if err != nil {
			s.ErrorLog.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - " + err.Error()))
			return
		}
	}

	w.Write([]byte("1"))
}

func (s *ServerHandler) Stop() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.InfoLog.Println("stop service")
	if err := s.WebServer.Shutdown(ctx); err != nil {
		s.InfoLog.Println("error stop")
		s.ErrorLog.Println(err)
	}
}

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"archistorage.unix/internal/archistorage/api"
	model "archistorage.unix/internal/archistorage/models"
)

const (
	fileConfig         = "config.ini"
	serviceDisplayName = "ArchiStorage"
	serviceName        = "ArchiStorage"
)

var server api.ServerHandler

var isRunning bool
var infoLog, errorLog *log.Logger

func runStorageService(inThread bool) {

	if isRunning {
		return
	}

	config := model.IniConfig{}

	infoLog.Println("Запуск сервиса ArchiStorage!")

	pathRoot, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		errorLog.Fatal(err)
	}

	config.DirConfig = pathRoot
	pathConfig := config.DirConfig + string(os.PathSeparator) + fileConfig

	infoLog.Println("Чтение файла конфигурации...", pathConfig)

	err = config.LoadFromFile(pathConfig)
	if err != nil {
		errorLog.Fatal(err.Error())
	}
	infoLog.Println("Файл конфигурации успешно прочитан!")

	isRunning = true

	server = api.ServerHandler{}
	server.Config = &config
	server.ErrorLog = errorLog
	server.InfoLog = infoLog
	server.Run(inThread)
}

func main() {

	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	path = path + string(os.PathSeparator) + "logs"
	err = os.MkdirAll(path, 0755)
	if err != nil {
		log.Fatal(err)
	}

	dt := time.Now()
	path = path + string(os.PathSeparator) + dt.Format("20060201150405") + ".log"

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	errorLog = log.New(f, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	infoLog = log.New(f, "INFO\t", log.Ldate|log.Ltime)

	infoLog.Println("Запуск ArchiStorage (файловое хранилище)")

	isRunning = false
	if len(os.Args) == 1 {
		fmt.Println("Запуск ArchiStorage (файловое хранилище)")
		runStorageService(false)
		return
	}

}

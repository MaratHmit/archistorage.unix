package db

import (
	"database/sql"
	_ "github.com/denisenkom/go-mssqldb"
	model "gitlab.archimed-soft.ru/web/archistorage/internal/archistorage/models"
	"log"
	"strings"
)

type Handler struct {
	Config   *model.IniConfig
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

func (h *Handler) getConnectString() string {
	password := strings.Replace(h.Config.DBPassword, "@", "%40", 1)
	return "sqlserver://" + h.Config.DBUser + ":" + password + "@" + h.Config.DBServer +
		"?database=" + h.Config.DBName + "&connection+timeout=30"
}

func (h *Handler) SaveFileInfo(infoF *model.FileInfo) int64 {
	connectString := h.getConnectString()
	db, err := sql.Open("mssql", connectString)
	if err != nil {
		h.ErrorLog.Println(err)
		return 0
	}
	defer db.Close()

	var id int64 = 0
	rows, err := db.Query("INSERT INTO STORAGE_FILE (NAME, EXTENSION, BOX_KEY, ORIGINAL_NAME, UPLOAD_DATETIME, UPLOADED_USER_ID) "+
		"VALUES (?, ?, ?, ?, ?, ?);  SELECT IDENT_CURRENT('STORAGE_FILE') ID",
		infoF.Name, infoF.Extension, infoF.BoxKey, infoF.OriginalName, infoF.UploadDateTime, infoF.UploadUserID)
	if err != nil {
		h.ErrorLog.Println(err)
		return 0
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&id)
	}

	return id
}

func (h *Handler) GetFileInfo(ID int, infoF *model.FileInfo) bool {
	connectString := h.getConnectString()
	db, err := sql.Open("mssql", connectString)
	if err != nil {
		h.ErrorLog.Println(err)
		return false
	}
	defer db.Close()

	row := db.QueryRow("SELECT NAME, EXTENSION, BOX_KEY, ORIGINAL_NAME, UPLOAD_DATETIME, UPLOADED_USER_ID FROM STORAGE_FILE WHERE ID = ?", ID)
	row.Scan(&infoF.Name, &infoF.Extension, &infoF.BoxKey, &infoF.OriginalName, &infoF.UploadDateTime, &infoF.UploadUserID)

	return true
}

func (h *Handler) GetFileName(id int) (string, error) {
	connectString := h.getConnectString()
	db, err := sql.Open("mssql", connectString)
	if err != nil {
		h.ErrorLog.Println(err)
		return "", err
	}
	defer db.Close()

	var name string
	row := db.QueryRow("SELECT NAME FROM STORAGE_FILE WHERE ID = ?", id)
	err = row.Scan(&name)
	if err != nil {
		h.ErrorLog.Println(err)
		return "", err
	}

	return name, nil
}

func (h *Handler) Delete(id int) error {
	connectString := h.getConnectString()
	db, err := sql.Open("mssql", connectString)
	if err != nil {
		h.ErrorLog.Println(err)
		return err
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM STORAGE_FILE WHERE ID = ?", id)
	if err != nil {
		h.ErrorLog.Println(err)
		return err
	}

	return nil
}

func (h *Handler) Rename(id int, newName string) error {

	connectString := h.getConnectString()
	db, err := sql.Open("mssql", connectString)
	if err != nil {
		h.ErrorLog.Println(err)
		return err
	}
	defer db.Close()

	_, err = db.Exec("UPDATE STORAGE_FILE SET ORIGINAL_NAME = ? WHERE ID = ?", newName, id)
	if err != nil {
		h.ErrorLog.Println(err)
		return err
	}

	return nil
}

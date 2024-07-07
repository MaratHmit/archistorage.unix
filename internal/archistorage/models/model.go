package model

import (
	"os"
	"strings"

	"gopkg.in/ini.v1"
)

const (
	DirMedcart = "MedcartPhoto"
)

type FileInfo struct {
	ID             int     `json:"id,omitempty"`
	Name           string  `json:"name,omitempty"`
	Extension      string  `json:",omitempty"`
	BoxKey         string  `json:"box_key"`
	OriginalName   string  `json:"original_name"`
	UploadDateTime float32 `json:"upload_datetime"`
	UploadUserID   int     `json:"uploaded_user_id"`
	Url            string  `json:",omitempty"`
	CreatedAt      int64   `json:"created_at"`
}

type IniConfig struct {
	DBServer        string
	DBUser          string
	DBPassword      string
	DBName          string
	DirCompany      string
	DirConfig       string
	DirCash         string
	ProgramDir      string
	MedcartPhotoDir string
}

func (c *IniConfig) LoadFromFile(filePath string) error {

	cfg, err := ini.Load(filePath)
	if err != nil {
		return err
	}

	c.DBServer = cfg.Section("Database").Key("Server").String()
	c.DBServer = strings.Replace(c.DBServer, ",", ":", 1)
	c.DBName = cfg.Section("Database").Key("Database").String()
	c.DBUser = cfg.Section("Database").Key("Login").String()
	c.DBPassword = cfg.Section("Database").Key("Password").String()
	c.ProgramDir = cfg.Section("Main").Key("Dir").String()
	c.MedcartPhotoDir = c.ProgramDir + string(os.PathSeparator) + DirMedcart

	err = os.MkdirAll(c.MedcartPhotoDir, 0777)
	if err != nil {
		return err
	}

	return nil
}

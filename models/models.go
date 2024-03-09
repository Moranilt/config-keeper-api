package models

import (
	"mime/multipart"

	"github.com/Moranilt/http-utils/tiny_errors"
)

type TestRequest struct {
	Firstname  string  `json:"fistname"`
	Lastname   string  `json:"lastname"`
	Patronymic *string `json:"patronymic"`
}

type TestResponse struct {
	ID string `json:"id"`
}

type FileRequest struct {
	Name        string                  `mapstructure:"name"`
	Files       []*multipart.FileHeader `mapstructure:"file"`
	OneMoreFile *multipart.FileHeader   `mapstructure:"one_more_file"`
}

type FileResponse struct {
	Name        string                  `mapstructure:"name" json:"name"`
	Files       []*multipart.FileHeader `mapstructure:"file[]" json:"files"`
	OneMoreFile *multipart.FileHeader   `mapstructure:"one_more_file" json:"one_more_file"`
}

const (
	_ = iota
	ERR_CODE_Database
	ERR_CODE_Marshal
	ERR_CODE_Redis
	ERR_CODE_RabbitMQ
	ERR_CODE_BodyRequired
)

const (
	ERR_BodyRequired = "required body is missing"
)

var (
	ERR_BodyRequiredTiny = tiny_errors.New(ERR_CODE_BodyRequired, tiny_errors.Message(ERR_BodyRequired))
)

package ecode

import (
	"encoding/json"
)

var errors map[int]string

const (
	Success           = -1
	ParseURLException = 100 + iota
	AlreadyDownloaded
	NoCopyright
	BuildPathException
	HTTPRequestException
	APIResponseException
	BuildFileException
	FileTransferException
)

func init() {
	errors = make(map[int]string)

	errors[Success] = "everything is ok"
	errors[ParseURLException] = "parse url exception"
	errors[AlreadyDownloaded] = "already downloaded"
	errors[NoCopyright] = "no copyright"
	errors[BuildPathException] = "build path exception"
	errors[HTTPRequestException] = "http request exception"
	errors[APIResponseException] = "api response exception"
	errors[BuildFileException] = "build file exception"
	errors[FileTransferException] = "file transfer exception"
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Cause   string `json:"cause,omitempty"`
}

func NewError(code int, cause string) Error {
	return Error{
		Code:    code,
		Message: errors[code],
		Cause:   cause,
	}
}

func Message(code int) string {
	return errors[code]
}

func (e Error) Error() string {
	return e.Message
}

func (e Error) toJsonString() string {
	b, _ := json.Marshal(e)
	return string(b)
}

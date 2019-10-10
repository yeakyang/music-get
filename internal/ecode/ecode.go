package ecode

var errors map[int]string

const (
	Success           = -1
	ParseURLException = 1000 + iota
	AlreadyDownloaded
	SongUnavailable
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
	errors[SongUnavailable] = "song unavailable"
	errors[BuildPathException] = "build path exception"
	errors[HTTPRequestException] = "http request exception"
	errors[APIResponseException] = "api response exception"
	errors[BuildFileException] = "build file exception"
	errors[FileTransferException] = "file transfer exception"
}

func Message(code int) string {
	return errors[code]
}

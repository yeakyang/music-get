package netease

import (
	"encoding/json"
	"github.com/winterssy/music-get/common"
	"net/http"
	urlpkg "net/url"
	"strings"
)

func post(url string, data interface{}) (*http.Response, error) {
	enc, _ := json.Marshal(data)
	params, encSecKey, err := Encrypt(enc)
	if err != nil {
		return nil, err
	}

	form := urlpkg.Values{}
	form.Set("params", params)
	form.Set("encSecKey", encSecKey)
	return common.Request("POST", url, nil, strings.NewReader(form.Encode()), common.NeteaseMusic)
}

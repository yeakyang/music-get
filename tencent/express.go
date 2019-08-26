package tencent

import (
	"encoding/json"
	"fmt"

	"github.com/winterssy/music-get/common"
)

const (
	MusicExpressAPI = "https://c.y.qq.com/base/fcgi-bin/fcg_music_express_mobile3.fcg"
)

type MusicExpress struct {
	Code int `json:"code"`
	Data struct {
		Items []struct {
			Filename string `json:"filename"`
			VKey     string `json:"vkey"`
		}
	} `json:"data"`
}

// func createGuid() string {
// 	r := rand.New(rand.NewSource(time.Now().UnixNano()))
// 	return strconv.Itoa(r.Intn(10000000000-1000000000) + 1000000000)
// }

func getVKey(guid string, songmid string, brCode string, ext string) (vkey string, filename string, err error) {
	query := map[string]string{
		"cid":      "205361747",
		"guid":     guid,
		"format":   "json",
		"songmid":  songmid,
		"filename": fmt.Sprintf("%s%s.%s", brCode, songmid, ext),
	}

	resp, err := common.Request("GET", MusicExpressAPI, query, nil, common.TencentMusic)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var m MusicExpress
	if err = json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return
	}

	if len(m.Data.Items) == 0 {
		err = fmt.Errorf("get vkey failed: %s", songmid)
		return
	}

	vkey = m.Data.Items[0].VKey
	filename = m.Data.Items[0].Filename
	return
}

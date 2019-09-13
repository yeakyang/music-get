package tencent

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/winterssy/music-get/common"
)

const (
	MusicExpressAPI = "https://u.y.qq.com/cgi-bin/musicu.fcg"
)

type MusicExpress struct {
	Code int `json:"code"`
	Req0 struct {
		Data struct {
			MidUrlInfo []struct {
				Vkey string `json:"vkey"`
			} `json:"midurlinfo"`
			TestFile2g string `json:"testfile2g"`
		} `json:"data"`
	} `json:"req0"`
}

// func createGuid() string {
// 	r := rand.New(rand.NewSource(time.Now().UnixNano()))
// 	return strconv.Itoa(r.Intn(10000000000-1000000000) + 1000000000)
// }

func getVkey(guid, songmid string) (vkey string, err error) {
	param := map[string]interface{}{
		"guid":      guid,
		"loginflag": 1,
		"songmid":   []string{songmid},
		"songtype":  []int{0},
		"uin":       "0",
		"platform":  "20",
	}
	req0 := map[string]interface{}{
		"module": "vkey.GetVkeyServer",
		"method": "CgiGetVkey",
		"param":  param,
	}
	data := map[string]interface{}{
		"req0": req0,
	}
	enc, err := json.Marshal(data)
	if err != nil {
		return
	}

	query := map[string]string{
		"data": string(enc),
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

	if len(m.Req0.Data.MidUrlInfo) == 0 || m.Req0.Data.MidUrlInfo[0].Vkey == "" {
		s := regexp.MustCompile("vkey=(\\w+)").FindStringSubmatch(m.Req0.Data.TestFile2g)
		if len(s) < 2 || s[1] == "" {
			err = fmt.Errorf("get vkey failed: %s", songmid)
			return
		}
		vkey = s[1]
		return
	}

	vkey = m.Req0.Data.MidUrlInfo[0].Vkey
	return
}

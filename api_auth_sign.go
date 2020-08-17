package getui

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

type SignParam struct {
	Sign      string `json:"sign"`
	Timestamp string `json:"timestamp"`
	AppKey    string `json:"appkey"`
}

//token
type SignResult struct {
	Result    string `json:"result"`
	AuthToken string `json:"auth_token"`
}

//获取Auth签名
//http://docs.getui.com/getui/server/rest/other_if/
func GetGeTuiToken(appID string, appKey string, masterSecret string) (*SignResult, error) {

	signStr, timestamp := Signature(appKey, masterSecret)

	param := &SignParam{
		Sign:      signStr,
		Timestamp: timestamp,
		AppKey:    appKey,
	}

	bodyByte, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}

	url := API_URL + appID + "/auth_sign"
	result, err := SendPost(url, "", bodyByte)
	if err != nil {
		return nil, err
	}

	tokenResult := new(SignResult)
	if err := json.Unmarshal([]byte(result), &tokenResult); err != nil {
		return nil, err
	}

	return tokenResult, nil
}

var (
	DefaultCacheToken = CacheToken{
		caches:        make(map[cacheTokenKey]cacheTokenVal),
		ExpireSeconds: 60 * 60,
	}
)

type CacheToken struct {
	sync.Mutex

	ExpireSeconds int64
	caches        map[cacheTokenKey]cacheTokenVal
}
type cacheTokenKey struct {
	appid     string
	appkey    string
	masterSec string
}
type cacheTokenVal struct {
	token      string
	updateTime int64
}

func (ct *CacheToken) GetToken(appid, appkey, masterSecret string) (token string, err error) {
	ct.Lock()
	defer ct.Unlock()

	key := cacheTokenKey{
		appid:     appid,
		appkey:    appkey,
		masterSec: masterSecret,
	}
	now := time.Now()
	cacheToken, ok := ct.caches[key]
	if ok && now.Unix()-cacheToken.updateTime < ct.ExpireSeconds {
		token = cacheToken.token
		return
	}
	gtToken, err := GetGeTuiToken(appid, appkey, masterSecret)
	if err != nil {
		err = fmt.Errorf("get getui token:%v", err)
		return
	}
	if strings.ToLower(gtToken.Result) != "ok" {
		err = fmt.Errorf("getui result:%v", gtToken.Result)
		return
	}
	cacheItem := cacheTokenVal{
		token:      gtToken.AuthToken,
		updateTime: now.Unix(),
	}
	ct.caches[key] = cacheItem
	token = gtToken.AuthToken
	return
}

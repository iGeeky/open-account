package net

import (
	// "fmt"
	"encoding/json"
	"time"

	"github.com/iGeeky/open-account/pkg/baselib/cache"
	"github.com/iGeeky/open-account/pkg/baselib/log"
)

type RedisHttpClient struct {
	redisCfg *cache.RedisConfig // redis配置.
	cache    *cache.RedisCache
}

type CachedResp struct {
	Resp
	URI        string
	UpdateTime int32 //unixtime
	Exptime    int32 //unixtime
}

func NewRedisHttpClient(redisCfg *cache.RedisConfig) (client *RedisHttpClient, err error) {
	if redisCfg.Exptime == 0 {
		redisCfg.Exptime = time.Hour * 24 * 2
	}

	cache, err1 := cache.NewRedisCache(redisCfg)
	if err1 != nil {
		err = err1
		return
	}

	client = &RedisHttpClient{redisCfg, cache}

	return
}

func (r *RedisHttpClient) HttpGetJson(uri string, headers map[string]string,
	timeout time.Duration, exptime time.Duration) *OkJson {
	key, ok := headers["X-Key"]
	if !ok {
		res := HttpGetJson(uri, headers, timeout)
		res.Cached = "miss"
		return res
	}

	resCached := &OkJson{}
	now := time.Now()
	bodyLen := 0
	str, err := r.cache.Get(key)
	defer writeDebugOKJson(now, resCached, &bodyLen)
	if err == nil && str != "" {
		var data CachedResp
		err = json.Unmarshal([]byte(str), &data)
		// exptime > now 没过期.
		if err == nil {
			resCached.StatusCode = data.StatusCode
			resCached.RawBody = data.RawBody
			resCached.Headers = data.Headers
			resCached.ReqDebug = data.ReqDebug
			resCached.Cached = "hit"
			_, ok = headers["X-UseCacheOnFail"]
			if !ok && data.Exptime > int32(now.Unix()) {
				return OkJsonParse(resCached)
			}
		}
	}

	res := HttpGetJson(uri, headers, timeout)

	if res.StatusCode == 200 {
		data := &CachedResp{}
		data.StatusCode = res.StatusCode
		data.RawBody = res.RawBody
		data.Headers = res.Headers
		data.ReqDebug = res.ReqDebug
		data.URI = uri
		data.UpdateTime = int32(now.Unix())
		data.Exptime = int32(now.Add(exptime).Unix())
		bt, _ := json.Marshal(data)
		str = string(bt)
		err = r.cache.Set(key, str, r.redisCfg.Exptime)
		if err != nil {
			log.Errorf("cache.Set('%s', '%v', '%v') failed! err:%v", key, data, r.redisCfg.Exptime, err)
		}
	} else {
		if resCached.StatusCode > 0 {
			return OkJsonParse(resCached)
		}
	}
	res.Cached = "miss"

	return res
}

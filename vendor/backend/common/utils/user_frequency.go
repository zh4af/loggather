/*
add by baixiao @ 2016-1-3
设定时间内，针对单个用户的操作不能超过限制
比如：当天对单个用户的的推送不超过2条，私信不超过10条
*/

package utils

import (
	"backend/common/clog"
	"backend/common/config"
	"backend/common/redisutil"
	"fmt"
	"regexp"
	"time"
)

const (
	CUR_DAY = -1
)

// expire为-1时，表示当天有效
// 判断用户在指定时间内是否能执行指定操作
func UserCmdValid(userId, cmd string, count int64, expire int64, redis *redisutil.Cache) bool {
	key := fmt.Sprintf("uf.%s.%s", userId, cmd)
	if expire == CUR_DAY {
		now := time.Now()
		year, month, day := now.Date()
		next := time.Date(year, month, day, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
		expire = int64(next.Sub(now) / time.Second) //当天剩余秒数
	}
	clog.Debugf("expire %d", expire)

	data, err := redis.GetInt64(key)
	if nil != err {
		err = redis.Set(key, 1, int(expire))
		if err != nil {
			clog.Warnf("[key:%s] set 1,%d error: %v", key, expire, err)
			return false
		}
		return true
	} else {
		if data >= count {
			clog.Debugf("[key:%s] %d >= %d", key, data, count)
			return false
		} else {
			ttl, err := redis.TTL(key)
			if err != nil {
				clog.Warnf("[key:%s] ttl error: %v", key, err)
				return false
			}
			//clog.Warnf("ttl %d", ttl)
			err = redis.Set(key, data+1, ttl)
			if err != nil {
				clog.Warnf("[key:%s] set %d,%d error: %v", key, data+1, ttl, err)
				return false
			}
			return true
		}
	}
}

type GrayTestStr struct {
	CmdMap map[string]CmdConfig
}

type CmdConfig struct {
	StartTime string
	EndTime   string
	Pattern   string
}

var GrayTest *GrayTestStr

// 判断用户能否执行指定模块的灰测
func UserGrayTest(userId, cmd string) bool {
	if GrayTest == nil || GrayTest.CmdMap == nil {
		addrs := []string{"http://etcdproxy.in.codoon.com:2381"}
		GrayTest = new(GrayTestStr)
		err := config.LoadCfgFromEtcd(addrs, "gray_test", GrayTest)
		if err != nil {
			clog.Warnf("init GrayTest from etcd err : %v", err)
			return false
		}
		fmt.Println(GrayTest)
	}

	if v, ok := GrayTest.CmdMap[cmd]; ok {
		start, err := time.ParseInLocation("2006-01-02 15:04:05", v.StartTime, time.Local) //todo 时间格式
		if err != nil {
			clog.Warnf("ParseInLocation %s error: %v", v.StartTime, err)
			return false
		}
		if time.Now().Before(start) {
			return false
		}
		end, err := time.ParseInLocation("2006-01-02 15:04:05", v.EndTime, time.Local) //todo 时间格式
		if err != nil {
			clog.Warnf("ParseInLocation %s error: %v", v.EndTime, err)
			return false
		}
		if time.Now().After(end) {
			return false
		}
		if match, err := regexp.MatchString(v.Pattern, userId); err != nil {
			clog.Warnf("MatchString %s,%s error: %v", v.Pattern, userId, err)
			return false
		} else {
			return match
		}
	}

	return false
}

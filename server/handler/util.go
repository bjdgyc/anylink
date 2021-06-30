package handler

import (
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/patrickmn/go-cache"
)

var ipcheck = cache.New(300*time.Second, 60*time.Second)
var usercheck = cache.New(300*time.Second, 60*time.Second)
var backtime = new(cache.Cache)
var ipnum int
var usernum int

func get_ipcheck() int {
	conf := base.ServerCfg2Slice()
	//var dbconfig string
	for _, j := range conf {
		if j.Name == "ip_check_num" {
			if config, ok := j.Data.(int); ok {
				return config
			}

		}
	}
	return 50
}

func get_usercheck() int {
	conf := base.ServerCfg2Slice()
	//var dbconfig string
	for _, j := range conf {
		if j.Name == "user_check_num" {
			if config, ok := j.Data.(int); ok {
				return config
			}

		}
	}
	return 10
}

func init_backtime() *cache.Cache {
	conf := base.ServerCfg2Slice()
	//var dbconfig string
	for _, j := range conf {
		if j.Name == "back_time" {
			if config, ok := j.Data.(int); ok {
				return cache.New(time.Duration(config)*time.Second, 60*time.Second)
			}

		}
	}
	return cache.New(300*time.Second, 60*time.Second)
}

func initBack() {

	backtime = init_backtime()
	ipnum = get_ipcheck()
	usernum = get_usercheck()
}

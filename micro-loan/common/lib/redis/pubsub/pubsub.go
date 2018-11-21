package pubsub

import (
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/gomodule/redigo/redis"
)

var redisHost string
var address string
var redisDb int
var redisPort int

func init() {
	redisHost = beego.AppConfig.String("cache_redis_host")
	redisPort, _ = beego.AppConfig.Int("cache_redis_port")
	redisDb, _ = beego.AppConfig.Int("cache_redis_db")
	address = fmt.Sprintf("%s:%d", redisHost, redisPort)
}

func Subscribe(ch string, c chan []byte) {
	for {
		err := startSubscribe(ch, c)
		logs.Error("[Subscribe] startSubscribe return error ch:%s, err:%v", ch, err)

		time.Sleep(time.Second)
	}
}

func createAndListenConn(ch string) (*redis.PubSubConn, error) {
	c, err := redis.Dial("tcp", address)
	if err != nil {
		logs.Error("[createAndListenConn] Dial return error ch:%s, address:%s, err:%v", ch, address, err)
		return nil, err
	}

	psc := new(redis.PubSubConn)
	psc.Conn = c
	err = psc.Subscribe(ch)
	if err != nil {
		logs.Error("[createAndListenConn] Subscribe return error ch:%s, err:%v", ch, err)
		return nil, err
	}

	return psc, nil
}

func startSubscribe(ch string, c chan []byte) error {
	psc, err := createAndListenConn(ch)
	if err != nil {
		return err
	}

	defer psc.Conn.Close()

	dataCh := make(chan []byte)
	errCh := make(chan error)
	go receiveMessage(psc, dataCh, errCh)

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for err == nil {
		select {
		case data := <-dataCh:
			c <- data
		case err = <-errCh:

		case <-ticker.C:
			if err := psc.Ping(""); err != nil {
				psc.Unsubscribe()
			}
		}
	}

	close(dataCh)
	close(errCh)

	return err
}

func receiveMessage(psc *redis.PubSubConn, dataCh chan []byte, errCh chan error) {
	for {
		switch t := psc.Receive().(type) {
		case error:
			errCh <- t
			logs.Error("[receiveMessage] receive error err:%v", t)
			return
		case redis.Message:
			dataCh <- t.Data
		case redis.Subscription:
			switch t.Count {
			case 0:
				errCh <- fmt.Errorf("unsubscribe")
				logs.Error("[receiveMessage] unsubscribe data:%v", t)
				return
			}
		case redis.Pong:

		default:
			logs.Warn("[receiveMessage] receive unhandle type data:%v", t)
		}
	}
}

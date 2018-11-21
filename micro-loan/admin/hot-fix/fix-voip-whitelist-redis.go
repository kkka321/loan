package main

import (
	"bufio"
	"fmt"
	"io"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/tools"
	"os"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"
)

func main() {
	// 设置进程 title
	procTitle := "fix-voip-whitelist-redis"
	gspt.SetProcTitle(procTitle)

	logs.Info("[%s] start launch.", procTitle)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := fmt.Sprintf("lock:%s", procTitle)
	storageClient.Do("DEL", lockKey)

	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")

	if err != nil || lock == nil {
		logs.Error("[%s] process is working, so, I will exit.", procTitle)
		return
	}

	setName := "set:voip-white-list"

	fileName := "./white_list.txt"
	file, err := os.OpenFile(fileName, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Open txt file error:", err)
		return
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	i := 0
	for {
		i++
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		mobile := mobileFormat(line)
		if err != nil {
			if err == io.EOF {
				fmt.Println("File read ok, i:", i, line)
				break
			} else {
				fmt.Println("Read file error:", err, "i:", i, line)
				return
			}
		}

		qVal, err := storageClient.Do("SADD", setName, mobile)
		// 说明有错,或已经处理过,忽略本次操作
		if err != nil || 0 == qVal.(int64) {
			logs.Info("[Voip-White-List-Redis] 次手机号已添加. mobile: %s", mobile)
		}

	}

	logs.Info("[%s] politeness exit.", procTitle)
	storageClient.Do("DEL", lockKey)
}

func mobileFormat(mobile string) string {
	if strings.HasPrefix(mobile, "628") {
		return strings.Replace(mobile, "628", "08", 1)
	}
	return mobile
}

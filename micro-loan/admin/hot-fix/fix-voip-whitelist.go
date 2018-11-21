package main

import (
	"bufio"
	"fmt"
	"io"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"os"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"
)

func main() {
	// 设置进程 title
	procTitle := "fix-voip-whitelist"
	gspt.SetProcTitle(procTitle)

	logs.Info("[%s] start launch.", procTitle)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := fmt.Sprintf("lock:%s", procTitle)
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")
	if err != nil || lock == nil {
		logs.Error("[%s] process is working, so, I will exit.", procTitle)
		return
	}

	voipWhite := models.VoipWhiteList{}

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

		voipWhite.Mobile = mobile

		_, err = voipWhite.Insert()
		if err == nil {
			//logs.Debug("----- voip 白名单插入成功：", voipWhite, "i:", i, "mobile:", mobile)
		} else {
			logs.Error("----- voip 白名单插入失败：", voipWhite, "1:", i, "mobile:", mobile)
			break
		}

		if i > 5 {
			break
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

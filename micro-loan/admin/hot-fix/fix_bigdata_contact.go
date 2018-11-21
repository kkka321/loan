package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	// 数据库初始化
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	"github.com/astaxie/beego/logs"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/service"
	"micro-loan/common/tools"
)

// 大数据通讯录导入 {{{
func main() {
	logs.Info("[fix_bigdata_contact] start launch.")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := "lock:fix_bigdata_contact"
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")

	if err != nil || lock == nil {
		logs.Error("[fix_bigdata_contact] process is working, so, I will exit.")
		return
	}
	// startDate := time.Now().AddDate(0, 0, -40).Format("2006-01-02")
	startDate := "2018-07-15"
	startDateTimestamp, _ := tools.GetTimeParseWithFormat(startDate, "2006-01-02")

	for i := 0; i < 100; i++ {
		datestr := tools.MDateMHSDate(startDateTimestamp*1000 + int64(86400*i*1000))
		logs.Debug("begin working ,date:", datestr)

		err := DoInsertBigdataContact(datestr)
		if err != nil {
			logs.Error("[fix_bigdata_contact] happend error:", err)
			break
		}
	}

	// for {

	// 	doInsertBigdataContact()
	// 	break
	// }
	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[BigdataContactTask] politeness exit.")
}

func DoInsertBigdataContact(date string) error {
	bucket, prefix := service.AwsBucketBigData()
	prefix = fmt.Sprintf(prefix, date)
	logs.Debug("bucket:%s prefix:%s", bucket, prefix)
	data, err := service.AwsListObjects(bucket, prefix)

	// logs.Notice("[data:]", data)

	if err != nil {
		return err
	}
	tmpPath := "/tmp/bigdata-contact/"
	for _, s3key := range data {

		file := tmpPath + service.GetFileName(s3key)
		logs.Debug("马上开始下载：", s3key)
		_, err := service.AwsDownloadByBucketAndKey(bucket, s3key, tmpPath, file)

		if err != nil {
			logs.Error("ERR:", err)
			break
		} else {
			logs.Debug("下载成功：", file)
			defer tools.Remove(file)

			f, err := os.Open(file)
			if err != nil {
				return err
			}
			defer f.Close()

			rd := bufio.NewReader(f)
			for {

				line, err := rd.ReadString('\n') //以'\n'为结束符读入一行
				if err != nil || io.EOF == err {
					break
				}

				// logs.Warning("line:", line)

				lineSlice := strings.Split(strings.Replace(line, "\n", "", 1), "\x01")
				if len(lineSlice) == 5 {
					// logs.Warning("lineSlice:", lineSlice)
					for range lineSlice {
						accountID, _ := tools.Str2Int64(lineSlice[0])
						mobile := lineSlice[1]
						contactName := lineSlice[2]
						ctime, _ := tools.Str2Int64(lineSlice[3])
						utime, _ := tools.Str2Int64(lineSlice[4])

						contact, _ := models.OneAccountBigdataContactByIM(accountID, mobile)
						if contact.Id == 0 {
							accountBigdataContact := models.AccountBigdataContact{
								AccountID:   accountID,
								Mobile:      mobile,
								ContactName: contactName,
								Ctime:       ctime,
								Utime:       utime,
								Itime:       tools.GetUnixMillis(),
								S3key:       s3key,
							}
							models.OrmInsert(&accountBigdataContact)
							// logs.Warning("数据插入")
						} else {
							// logs.Error("数据已存在跳过")
						}
					}
				}
			}
		}
	}
	return err
}

// }}}

package task

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"micro-loan/common/lib/gaws"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
)

type BigdataContactTask struct {
}

// 大数据通讯录导入 {{{
func (c *BigdataContactTask) Start() {
	logs.Info("[TaskHandleRepayRemindOrder] start launch.")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := beego.AppConfig.String("bigdata_contact")
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")

	if err != nil || lock == nil {
		logs.Error("[BigdataContactTask] process is working, so, I will exit.")
		return
	}
	for {
		if cancelled() {
			logs.Info("[TaskHandleRepayRemindOrder] receive exit cmd.")
			break
		}
		doInsertBigdataContact()
		break
	}
	// -1 正常退出时,释放锁
	storageClient.Do("DEL", lockKey)
	logs.Info("[BigdataContactTask] politeness exit.")
}
func (c *BigdataContactTask) Cancel() {
	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	lockKey := beego.AppConfig.String("bigdata_contact")
	storageClient.Do("DEL", lockKey)
}

func doInsertBigdataContact() {
	bucket, prefix := gaws.AwsBucketBigData()

	yesterday := tools.MDateMHSDate(tools.GetUnixMillis() - 86400*2*1000)

	prefix = fmt.Sprintf(prefix, yesterday)

	logs.Debug("bucket:%s prefix:%s", bucket, prefix)

	data, err := gaws.AwsListObjects(bucket, prefix)

	if err != nil {
		return
	}
	tmpPath := "/tmp/bigdata-contact/"
	for _, s3key := range data {

		file := tmpPath + gaws.GetFileName(s3key)
		logs.Debug("马上开始下载：", s3key)
		_, err := gaws.AwsDownloadByBucketAndKey(bucket, s3key, tmpPath, file)

		if err != nil {
			logs.Error("ERR:", err)
			break
		} else {
			logs.Debug("下载成功：", file)
			defer tools.Remove(file)

			f, err := os.Open(file)
			if err != nil {
				return
			}
			defer f.Close()

			rd := bufio.NewReader(f)
			for {

				line, err := rd.ReadString('\n') //以'\n'为结束符读入一行
				if err != nil || io.EOF == err {
					break
				}

				logs.Warning("line:", line, "file:", file)

				lineSlice := strings.Split(strings.Replace(line, "\n", "", 1), "\x01")
				logs.Debug("lineSliceLen:", len(lineSlice))
				if len(lineSlice) == 5 {
					logs.Warning("lineSlice:", lineSlice)
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
							logs.Warning("数据插入")
						} else {
							logs.Error("数据已存在跳过")
						}
					}
				}
			}
		}
	}
}

// }}}

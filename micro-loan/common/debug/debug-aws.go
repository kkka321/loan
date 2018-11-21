package main

import (
	"bufio"
	"fmt"
	"io"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"os"
	"strings"

	"github.com/astaxie/beego/logs"

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/service"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func main() {
	//err := upload()
	//fmt.Printf("%v\n", err)

	bucket, prefix := service.AwsBucketBigData()

	yesterday := tools.MDateMHSDate(tools.GetUnixMillis() - 86400*1000)

	prefix = fmt.Sprintf(prefix, yesterday)

	logs.Debug("bucket:%s prefix:%s", bucket, prefix)

	data, err := service.AwsListObjects(bucket, prefix)

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

				logs.Warning("line:", line)

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

	logs.Debug("data:", data, "err", err)
	os.Exit(0)

	num, err := service.AwsDownload("dev/README.md", "/tmp/test.md")
	fmt.Printf("num: %d, err: %v\n", num, err)
}

func upload() error {
	// The session the S3 Uploader will use
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-1"),
	}))

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	filename := "./README.md"
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", filename, err)
	}

	// Upload the file to S3.
	myBucket := "mobimagic-microl"
	myString := "dev/README.md"
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(myBucket),
		Key:    aws.String(myString),
		Body:   f,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}

	fmt.Printf("file uploaded to, %s\n", result.Location)

	return nil
}

package gaws

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"micro-loan/common/tools"
	"os"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// AwsBucket s3 桶名
//const AwsBucket string = "mobimagic-microl"

func AwsBucket() string {
	return beego.AppConfig.String("aws_s3_bucket")
}

func AdsBucket() string {
	return beego.AppConfig.String("ad_s3_bucket")
}

// AwsBucketBigData 大数据通讯录bucket & prefix
func AwsBucketBigData() (bucket, prefix string) {
	bucket = beego.AppConfig.String("aws_bigdata_contact_bucket")
	prefix = beego.AppConfig.String("aws_bigdata_contact_prefix")
	return
}

func AwsRiskBucket() string {
	return beego.AppConfig.String("aws_s3_risk_bucket")
}

// AwsResouceDomain aws资源域名
const AwsResouceDomain string = "https://d21sf181boh67y.cloudfront.net"

func createAwsSession() *session.Session {
	// The session the S3 Uploader will use
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-1"),
	}))

	return sess
}

// AwsUpload ...
// localFilename: 待上传的文件,可以是相对或绝对路径
// s3Key: 上传到s3的文件别名
func AwsUpload(localFilename, s3Key string) (resultLocation string, err error) {
	sess := createAwsSession()
	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	f, err := os.Open(localFilename)
	if err != nil {
		err = fmt.Errorf("failed to open file %q, %v", localFilename, err)
		return
	}

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(AwsBucket()),
		Key:    aws.String(s3Key),
		Body:   f,
	})
	if err != nil {
		err = fmt.Errorf("failed to upload file, %v", err)
		return
	}

	resultLocation = result.Location

	return
}

func AwsDelete(s3Key string) error {
	sess := createAwsSession()

	batcher := s3manager.NewBatchDelete(sess)

	objects := []s3manager.BatchDeleteObject{
		{
			Object: &s3.DeleteObjectInput{
				Key:    aws.String(s3Key),
				Bucket: aws.String(AwsBucket()),
			},
		},
	}

	err := batcher.Delete(aws.BackgroundContext(), &s3manager.DeleteObjectsIterator{
		Objects: objects,
	})

	return err
}

func AwsUploadStream(s3Key string, r io.Reader) (resultLocation string, err error) {
	sess := createAwsSession()

	uploader := s3manager.NewUploader(sess)

	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(AwsBucket()),
		Key:    aws.String(s3Key),
		Body:   r,
	})
	if err != nil {
		err = fmt.Errorf("failed to upload file, %v", err)
		return
	}

	resultLocation = result.Location

	return
}

func AdsUpload(localFilename, s3Key string) (resultLocation string, err error) {
	sess := createAwsSession()
	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	f, err := os.Open(localFilename)
	if err != nil {
		err = fmt.Errorf("failed to open file %q, %v", localFilename, err)
		return
	}

	// Upload the file to S3.
	logs.Error("-----------------------bucket: ", aws.String(AdsBucket()), aws.String(s3Key))
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(AdsBucket()),
		Key:    aws.String(s3Key),
		Body:   f,
	})
	if err != nil {
		err = fmt.Errorf("failed to upload file, %v", err)
		return
	}

	resultLocation = result.Location

	return
}

// BootLogUpload 上传启动日志到风控s3
func BootLogUpload(clientData string) (resultLocation string, err error) {
	//
	serverTime := tools.GetUnixMillis()
	pushMap := make(map[string]interface{})
	pushMap["t"] = tools.GetUnixMillis()
	pushMap["data"] = clientData
	logs.Debug("[BootLogUpload] upload map:", pushMap)
	pushBytes, _ := json.Marshal(pushMap)

	md5Ctx := md5.New()
	md5Ctx.Write(pushBytes)
	contentMd5 := hex.EncodeToString(md5Ctx.Sum(nil))

	fileName := strconv.FormatInt(serverTime, 10) + "-" + contentMd5 + "." + "json"
	saveKey := "boot-log/" + tools.GetToday() + "/" + fileName
	logs.Debug("[BootLogUpload] save key:", saveKey)
	sess := createAwsSession()
	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)
	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(AwsBucket()),
		Key:    aws.String(saveKey),
		Body:   bytes.NewReader(pushBytes),
	})
	if err != nil {
		err = fmt.Errorf("[BootLogUpload] failed to upload boot data, %v, clientData: %s", err, clientData)
		return
	}

	resultLocation = result.Location

	return
}

func AdUpload(localFilename, s3Key string) (resultLocation string, err error) {
	sess := createAwsSession()
	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	f, err := os.Open(localFilename)
	if err != nil {
		err = fmt.Errorf("failed to open file %q, %v", localFilename, err)
		return
	}

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(AwsBucket()),
		Key:    aws.String(s3Key),
		Body:   f,
	})
	if err != nil {
		err = fmt.Errorf("failed to upload file, %v", err)
		return
	}

	resultLocation = result.Location

	return
}

// AwsDownload 将s3对应的文件下载到本地指定的文件
func AwsDownload(s3Key, localFilename string) (num int64, err error) {
	sess := createAwsSession()
	// Create a downloader with the session and default options
	downloader := s3manager.NewDownloader(sess)

	// Create a file to write the S3 Object contents to.
	f, err := os.Create(localFilename)
	if err != nil {
		err = fmt.Errorf("failed to create file %q, %v", localFilename, err)
		return
	}

	// Write the contents of S3 Object to the file
	num, err = downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(AwsBucket()),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		err = fmt.Errorf("failed to download file, %v", err)
		return
	}

	return
}

// AwsDownload 将s3对应的文件写入对缓冲区
func AwsDownload2Stream(s3Key string, w io.WriterAt) (num int64, err error) {
	sess := createAwsSession()
	// Create a downloader with the session and default options
	downloader := s3manager.NewDownloader(sess)

	// Write the contents of S3 Object to stream
	num, err = downloader.Download(w, &s3.GetObjectInput{
		Bucket: aws.String(AwsBucket()),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		err = fmt.Errorf("failed to download file to buffer, %v", err)
		return
	}

	return
}

func AwsDownloadAdStream(s3Key string, w io.WriterAt) (num int64, err error) {
	sess := createAwsSession()
	// Create a downloader with the session and default options
	downloader := s3manager.NewDownloader(sess)

	// Write the contents of S3 Object to stream
	num, err = downloader.Download(w, &s3.GetObjectInput{
		Bucket: aws.String(AdsBucket()),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		err = fmt.Errorf("failed to download file to buffer, %v", err)
		return
	}

	return
}

/* AwsListObjects 列出文件
*  完整路径：s3://mobimagic-microl-risk/risk_base/dm_risk_base_user_contact_list_dt/dt=2018-07-21/000001_0
*  bucket:  mobimagic-microl-risk
*  prefix:  risk_base/dm_risk_base_user_contact_list_dt/dt=2018-07-21
*  return [000001_0,000002_0]
 */
func AwsListObjects(bucket, prefix string) (data []string, err error) {
	sess := createAwsSession()
	svc := s3.New(sess)
	err = svc.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}, func(p *s3.ListObjectsOutput, last bool) (shouldContinue bool) {
		count := len(p.Contents)
		for k, obj := range p.Contents {
			if k+1 <= count {
				data = append(data, string(*obj.Key))
			}
		}
		return
	})
	return
}

// AwsDownloadByBucketAndKey 将s3对应的文件下载到本地指定的文件
func AwsDownloadByBucketAndKey(bucket, s3Key, savedir, localFilename string) (num int64, err error) {
	sess := createAwsSession()
	// Create a downloader with the session and default options
	downloader := s3manager.NewDownloader(sess)

	mkdirerr := os.MkdirAll(savedir, 0777)
	if mkdirerr != nil {
		err = fmt.Errorf("MkdirAll error %v", mkdirerr)
		return
	}
	// Create a file to write the S3 Object contents to.
	f, err := os.Create(localFilename)
	if err != nil {
		err = fmt.Errorf("failed to create file %q, %v", localFilename, err)
		return
	}
	// Write the contents of S3 Object to the file
	num, err = downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		err = fmt.Errorf("failed to download file, %v", err)
		return
	}

	return
}

func buildTmpFilename(rid int64) (tmpFilename string) {
	return fmt.Sprintf("/tmp/%d.jpeg", rid)
}

// BuildTmpFilename 根据资源ID创建临时文件
func BuildTmpFilename(rid int64) (tmpFilename string) {
	return buildTmpFilename(rid)
}

// GetFileName 获取AWS完整路径中的文件名
func GetFileName(path string) (name string) {
	pathSlice := strings.Split(path, "/")
	name = pathSlice[len(pathSlice)-1]
	return
}

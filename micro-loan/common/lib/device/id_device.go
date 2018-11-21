package device

import (
	"fmt"
	"strconv"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

/**
生成形如: 180120981234561234,
0-5:  	YYMMDD
6-7:  	BizSN
8-13:	device seq id
14-17:  毫秒数最后4位

2018.08.16 升级
生成形如: 180102981234567812,
0-5:  	YYMMDD
6-7:  	BizSN
8-15:	device seq id
16-17:  毫秒数最后2位
*/
func GenerateBizId(bizSN types.BizSN) (int64, error) {
	t := time.Now()
	nanos := t.UnixNano()
	millis := nanos / 1000000

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	idDeviceKey := beego.AppConfig.String("id_device_key_hash")

	field := tools.GetDate(millis / 1000)
	id, _ := storageClient.Do("HINCRBY", idDeviceKey, field, 1)

	logs.Info("[GenerateBizId] millis:%d field :%s id:%d", millis, field, id)
	bizIdStr := fmt.Sprintf("%d%02d%02d%02d%08d%02d", t.Year()%100, t.Month(), t.Day(), bizSN, id.(int64)%100000000, millis%100)
	bizId, err := strconv.ParseInt(bizIdStr, 10, 64)

	return bizId, err
}

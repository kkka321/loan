package ticket

import (
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/gomodule/redigo/redis"
)

// 员工上线状态管理
// 员工当日首次登录, 将员工置为 上线状态
// worker_online_prefix:{adminUID}
// EXPIRE worker_online_prefix:{adminUID} 23:30timestamp - nowtimestamp
// 使用 redis

// 为减少定时任务
// 此处用 redis set 设计,
// 首次有人登录时, 若没有在线池, 则set 然后设置过期时间为EXPIREAT 23:30timestamp

// WorkerLogin 用户上线
func WorkerLogin(adminUID, roleID, lastLoginTime int64) {
	if todayStartTime, _ := tools.GetTodayTimestampByLocalTime("00:00"); lastLoginTime > todayStartTime*1000 {
		// 不是今天首次登录
		// 不做任何操作
		logs.Debug("[todayStartTime]not day first login, ignore, lastLoginTime:", lastLoginTime, "; today start Time:", todayStartTime*1000)
		return
	}
	// 首次登录,
	isNewOnlineWorker := doWorkerOnline(adminUID, roleID, adminUID)
	if isNewOnlineWorker {
		// 首次登录用户

		event.Trigger(&evtypes.WorkerDailyFirstOnlineEv{AdminUID: adminUID, RoleID: roleID})
	}
	return
}

// 此处的登录可能不是首次登录
// 可能为已经登录, 然后被下线,
func doWorkerOnline(adminUID, roleID, opUID int64) (isNewOnlineWorker bool) {
	// 判断时间
	// 是否是在 晚上11:30 之前
	expireTime := config.ValidItemString("ticket_worker_force_offline_time")
	expireTimestamp, _ := tools.GetTodayTimestampByLocalTime(expireTime)
	if time.Now().Unix() > expireTimestamp {
		// 下班时间之后登录, 不触发工作在线
		logs.Debug("[doWorkerOnline] login occur on work force_offline_time:", expireTimestamp, ";now:", time.Now().Unix())
		return
	}

	// 先检查用户work_status , 若为正常则进入在线池
	admin, _ := models.OneAdminByUid(adminUID)
	if admin.Status != types.StatusValid || admin.WorkStatus != types.AdminWorkStatusNormal {
		// 用户状态不符合条件, 不入在线工作池
		// 此时用户可能在休假, 只是临时登录解决某些问题
		return
	}

	onlineSetKey := beego.AppConfig.String("worker_online_sets")
	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()

	isExists, _ := redis.Bool(redisCli.Do("EXISTS", onlineSetKey))

	// isExists, _ := setIsExists(redisCli, onlineSetKey)
	// 若不存在, 则在 SADD 成功后设置过期时间

	count, _ := redis.Int(redisCli.Do("SADD", onlineSetKey, adminUID))
	if count > 0 {
		// online success
		logs.Debug("[doWorkerOnline] is new daily online user:", adminUID)
		models.OpLogWrite(opUID, adminUID, models.OpCodeWorkerOnlineStatusUpdate, "worker_online_status",
			map[string]bool{"online": false}, map[string]bool{"online": true})
		isNewOnlineWorker = true
		PollWatchRoleOnlineUser(roleID, adminUID)

		if !isExists {
			redisCli.Do("EXPIREAT", onlineSetKey, expireTimestamp)
		}
	}

	return
}

// doWorkerOffline 用户下线
// 手动下线, 如何避免,在登录时触发上线, 只有首次登录触发 UserOnline, 利用lastLoginTime
func doWorkerOffline(adminUID, opUID int64) {
	// 先检查用户work_status , 若为正常则进入在线池
	if !isWorkerOnline(adminUID) {
		return
	}

	onlineSetKey := beego.AppConfig.String("worker_online_sets")
	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()

	affected, err := redis.Int(redisCli.Do("SREM", onlineSetKey, adminUID))
	if err != nil {
		logs.Error("[WorkerOffline] redis err:", err)
		return
	}

	if affected == 1 {
		// offline sccuess
		models.OpLogWrite(opUID, adminUID, models.OpCodeWorkerOnlineStatusUpdate, "worker_online_status",
			map[string]bool{"online": true}, map[string]bool{"online": false})
	}
}

// ManualWorkerOffline 手动下线, 停止接单
func ManualWorkerOffline(adminUID, opUID int64) {
	admin, _ := models.OneAdminByUid(adminUID)
	doWorkerOffline(adminUID, opUID)

	PollWatchRoleOfflineUser(admin.RoleID, adminUID)
}

// ManualWorkerOnline 手动下线, 停止接单
func ManualWorkerOnline(adminUID, opUID int64) {
	admin, _ := models.OneAdminByUid(adminUID)
	doWorkerOnline(adminUID, admin.RoleID, opUID)
}

// IsWorkerOnline 用户是否在线
func IsWorkerOnline(adminUID int64) bool {
	return isWorkerOnline(adminUID)
}

// isWorkerOnline 用户是否在线
func isWorkerOnline(adminUID int64) bool {
	// 先检查用户work_status , 若为正常则进入在线池
	onlineSetKey := beego.AppConfig.String("worker_online_sets")
	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()

	isMember, err := redis.Bool(redisCli.Do("SISMEMBER", onlineSetKey, adminUID))
	if err != nil {
		logs.Error("[IsWorkerOnline] redis err:", err)
	}

	return isMember
}

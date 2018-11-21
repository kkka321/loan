package ticket

import (
	"fmt"
	"math/rand"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/gomodule/redigo/redis"
)

// oneTicketDefaultWorkload 一个工单的负载值
const oneTicketDefaultWorkloadFactor int64 = 100

// WorkerAssignStrategy 人力分配策略
type WorkerAssignStrategy interface {
	OneWorker() (adminUID int64, err error)
	UserStartWork(adminUID, roleID int64)
	UserStopWork(adminUID, roleID int64)
}

// IdleWorkerStrategy 空闲员工策略
// 最闲的员工优先获取ticket
type IdleWorkerStrategy struct {
	TicketItem types.TicketItemEnum
}

// OneWorker 获取一个工单可分配工人
func (s *IdleWorkerStrategy) OneWorker() (adminUID int64, err error) {
	return s.idleAssign()
}

// idleAssign 空闲分配
func (s *IdleWorkerStrategy) idleAssign() (adminUID int64, err error) {
	// 最大同时工作量
	//maxWorkingTicket := config.ValidItemInt64("ticket_max_working_num_for_")
	var maxWorkingTicket int64
	if s.TicketItem == types.TicketItemPhoneVerify || s.TicketItem == types.TicketItemInfoReview {
		maxWorkingTicket, err = config.ValidItemInt64("ticket_max_working_num_for_phone_verify")
		if err != nil {
			maxWorkingTicket = 20
		}
	}
	maxWorkload := maxWorkingTicket * oneTicketDefaultWorkloadFactor

	key, err := s.getTicketItemIncompletedSetKey()
	if err != nil {
		logs.Error(err, "[idleAssign][getTicketItemIncompletedSetKey] err ticketItem:", s.TicketItem)
		return
	}
	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()
	//
	if !s.isPoolInit(key) {
		s.initPool()
	}

	// adminUID, _ := redis.Int64(redisCli.Do("ZRANGEBYSCORE", key, "-inf", maxWorkingTicket, "LIMIT", 0, 1))
	// score 值最小的用户
	logs.Debug("key", key)
	mostIdleWorkerWithScore, errMinScore := redis.Int64s(redisCli.Do("ZRANGEBYSCORE", key, "-inf", "+inf", "WITHSCORES", "LIMIT", 0, 1))
	logs.Debug("uid Score:", mostIdleWorkerWithScore)
	logs.Debug("uid Score:%T", mostIdleWorkerWithScore)
	if errMinScore != nil {
		err = fmt.Errorf("[idleAssign] redis err: %v", errMinScore)
		return
	}
	if len(mostIdleWorkerWithScore) < 2 {
		err = fmt.Errorf("[idleAssign]IdleWorkerWithScore Set is empty, no most idle worker, key: %s", key)
		return
	}

	// uid := mostIdleWorkerWithScore[0]
	if ticketNum := mostIdleWorkerWithScore[1]; ticketNum >= maxWorkload {
		logs.Debug("[idleAssign] most idle worker's ticket num(%d) >= maxWorkingTicket(%d)", ticketNum, maxWorkingTicket)
		// assign should wait and loop , until idle worker enough limit
		return
	}
	adminUID = mostIdleWorkerWithScore[0]

	// mostIdleWorkers, err2 := redis.Int64s(redisCli.Do("ZRANGEBYSCORE",
	// 	key, "-inf", "("+strconv.FormatInt(maxWorkingTicket, 10), "LIMIT", 0, 1))
	// if err2 != nil {
	// 	err = fmt.Errorf("[idleAssign] redis err: %v", errMinScore)
	// 	return
	// }
	// if len(mostIdleWorkers) > 0 {
	// 	adminUID = mostIdleWorkers[0]
	// }
	return
}

// 此方法由之前的，增量非幂等方法，
// 改为， 获取指定用户负载，重设（幂等方法），来减少逻辑复杂度，增加容错
// 相应，多了一次 sql 查询，微弱的性能损耗
func (s *IdleWorkerStrategy) workingTicketChange(adminUID int64) {
	key, _ := s.getTicketItemIncompletedSetKey()

	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()

	nowScore, err := redis.Int(redisCli.Do("ZSCORE", key, adminUID))
	if nowScore < 0 {
		// 按照目前的逻辑来看， nowScore 是工作负载
		logs.Error("[IdleWorkerStrategy-workingTicketChange]should not happen negative score: %d, admin uid: %d ",
			nowScore, adminUID)
	}
	// 若不存在，则会直接跳过， 不存在是 err = redis.ErrNil
	if err == nil {
		_, err := redisCli.Do("ZADD", key, s.calculateScore(adminUID), adminUID)

		//newScore, err := redis.Int(redisCli.Do("ZINCRBY", key, num, adminUID))
		//logs.Debug("[workingTicketChange] worker new score:", newScore)
		if err != nil {
			logs.Error("[workingTicketChange] ZADD error:", err)
		}
	}
}

// WorkingTicketChange 工作负载变化，重设负载
func (s *IdleWorkerStrategy) WorkingTicketChange(adminUID int64) {
	s.workingTicketChange(adminUID)
}

// UserStartWork 人力池监控用户上线
// 用于给角色新增用户,或者后台用户请假归来, 或者以后轮班
func (s *IdleWorkerStrategy) UserStartWork(adminUID, roleID int64) {
	// 不直接删除 queue重建, 是为维持, 当前分配序列, 确保,高频修改下, 分配均匀
	if !IsWorkerOnline(adminUID) {
		logs.Error("[IdleWorkerStrategy.UserStartWork] worker(%d) not online", adminUID)
		return
	}

	//
	key, _ := s.getTicketItemIncompletedSetKey()
	logs.Debug("[UserStartWork] adminUID(%d) ,roleID(%d)", adminUID, roleID)
	logs.Debug("[UserStartWork] pool key (%s)", key)
	if s.isPoolInit(key) {
		redisCli := storage.RedisStorageClient.Get()
		defer redisCli.Close()
		score := s.calculateScore(adminUID)
		_, err := redisCli.Do("ZADD", key, score, adminUID)
		if err != nil {
			logs.Error("[IdleWorkerStrategy.UserStartWork] redis err:", err)
		}
	}
}

func (s *IdleWorkerStrategy) calculateScore(adminUID int64) (score int64) {
	partialNum, workingNumWithOutPartial, err := models.GetWorkerIncompletedTicketNumByItem(adminUID, s.TicketItem)
	if err != nil {
		score = 999999
		return
	}
	score = s.partialWorkloadFactor()*partialNum + workingNumWithOutPartial*oneTicketDefaultWorkloadFactor

	return
}

func (s *IdleWorkerStrategy) partialWorkloadFactor() (factor int64) {
	factor = oneTicketDefaultWorkloadFactor / 2
	configFactor, err := config.ValidItemFloat64("ticket_partial_completed_workload_factor")
	if err != nil || configFactor < 0 || configFactor > 1 {
		logs.Error("[partialWorkloadFactor] err , not in range [0,1]:", err, configFactor)
		return
	}
	factor = int64(configFactor * float64(oneTicketDefaultWorkloadFactor))
	return
}

// UserStopWork 人力池监控用户上线
// 用于给角色新增用户,或者后台用户请假归来, 或者以后轮班
func (s *IdleWorkerStrategy) UserStopWork(adminUID, roleID int64) {
	// 不直接删除 queue重建, 是为维持, 当前分配序列, 确保,高频修改下, 分配均匀
	key, _ := s.getTicketItemIncompletedSetKey()
	if s.isPoolInit(key) {
		redisCli := storage.RedisStorageClient.Get()
		defer redisCli.Close()
		_, err := redisCli.Do("ZREM", key, adminUID)
		if err != nil {
			logs.Error("[IdleWorkerStrategy.UserStopWork] redis err:", err)
		}
		// 危险操作
	}
}

func (s *IdleWorkerStrategy) getMatchRuleAssignQueuePollByRole(roleID int64) (string, error) {
	keyPrefix := s.getAssignPollKeyPrefix()

	if len(keyPrefix) <= 5 {
		return "", fmt.Errorf("[getMatchAssignQueuePollByTicketItem]Fuzzy matching key prefix  is too short , not secure, key rule:%s", keyPrefix)
	}
	itemKeyPrefix := keyPrefix + "T" + strconv.Itoa(int(s.TicketItem)) + "_*" + createAssignQueueNameRolePart(strconv.FormatInt(roleID, 10)) + "*"
	return itemKeyPrefix, nil
}

func (s *IdleWorkerStrategy) isPoolInit(key string) bool {
	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()
	isExists, _ := redis.Bool(redisCli.Do("EXISTS", key))
	return isExists
}

func (s *IdleWorkerStrategy) initPool() {
	admins := activeCanAssignUsersByTicketItem(s.TicketItem)

	if len(admins) == 0 {
		logs.Debug("[IdleWorkerStrategy.init] no active worker to init assign poll")
		return
	}
	key, err := s.getTicketItemIncompletedSetKey()
	if err != nil {
		return
	}

	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()

	//incompletedTicketsNum := map[int64]int64{}

	for _, admin := range admins {
		_, err := redisCli.Do("ZADD", key, s.calculateScore(admin.Id), admin.Id)
		if err != nil {
			logs.Error("[WorkerIncompletedTicketsByTicketItem] ZADD", err)
		}
	}

	//  设置过期
	expireTime := config.ValidItemString("ticket_worker_force_offline_time")
	expireTimestamp, _ := tools.GetTodayTimestampByLocalTime(expireTime)
	redisCli.Do("EXPIREAT", key, expireTimestamp)

	return
}

func (s *IdleWorkerStrategy) getAssignPollKeyPrefix() string {
	return beego.AppConfig.String("ticket_active_worker_incompleted_num_prefix")
}

// 获取分配池名称(redis queue name)
func (s *IdleWorkerStrategy) getTicketItemIncompletedSetKey() (key string, err error) {
	keyPrefix := s.getAssignPollKeyPrefix()
	logs.Debug("keyPrefix:", keyPrefix)
	idstrings, err := canAssignRoles(s.TicketItem)
	if err != nil {
		// err was already logged
		return
	}

	itemKeyPrefix := keyPrefix + createAssignQueueNameItemPart(s.TicketItem)
	key = itemKeyPrefix + createAssignQueueNameRolePart(idstrings...)
	return
}

// PollingWorkerStrategy 轮询人力策略
// 目前这个为旧的人力策略, 已弃用
type PollingWorkerStrategy struct {
	TicketItem types.TicketItemEnum
}

// OneWorker 从人力池获取一个人力
func (s *PollingWorkerStrategy) OneWorker() (adminUID int64, err error) {
	return s.pollAssign()
}

// UserStartWork 人力池监控用户上线
// 用于给角色新增用户,或者后台用户请假归来, 或者以后轮班
func (s *PollingWorkerStrategy) UserStartWork(adminUID, roleID int64) {
	// 不直接删除 queue重建, 是为维持, 当前分配序列, 确保,高频修改下, 分配均匀
	matchRule, err := s.getMatchRuleAssignQueuePollByRole(roleID)

	if err != nil {
		logs.Error(err)
		return
	}
	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()
	// 低效率操作, 会有过于频繁会降低redis的性能, 未实测
	// TODO 使用SET 保存对应关系, 来减少模糊匹配带来的性能影响
	redisKeys, _ := redisCli.Do("KEYS", matchRule)
	if matchQueues, ok := redisKeys.([]interface{}); ok && len(matchQueues) > 0 {
		for _, v := range matchQueues {
			if queueName, ok := v.([]byte); ok {
				redisCli.Do("LPUSH", string(queueName), adminUID)
			}
			// 危险操作
		}
	}
}

// UserStopWork 人力池监控用户上线
// 用于给角色新增用户,或者后台用户请假归来, 或者以后轮班
func (s *PollingWorkerStrategy) UserStopWork(adminUID, roleID int64) {
	// 不直接删除 queue重建, 是为维持, 当前分配序列, 确保,高频修改下, 分配均匀
	matchRule, err := s.getMatchRuleAssignQueuePollByRole(roleID)

	if err != nil {
		logs.Error(err)
		return
	}
	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()
	// 低效率操作, 会有过于频繁会降低redis的性能
	// TODO 使用SET 保存对应关系, 来减少模糊匹配带来的性能影响
	redisKeys, _ := redisCli.Do("KEYS", matchRule)
	if matchQueues, ok := redisKeys.([]interface{}); ok && len(matchQueues) > 0 {
		for _, v := range matchQueues {
			if queueName, ok := v.([]byte); ok {
				redisCli.Do("LREM", string(queueName), 0, adminUID)
			}
			// 危险操作
		}
	}
}

// pollAssign 轮询分配
func (s *PollingWorkerStrategy) pollAssign() (adminUID int64, err error) {
	key, err := s.getAssignPollKey()
	if err != nil {
		logs.Error(err, "[pollAssign][getAssignPollKey] err ticketItem:", s.TicketItem)
		return
	}
	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()

	out, err := redisCli.Do("RPOPLPUSH", key, key)
	if out == nil {
		// 如果为空, 则说明配置改动, 或者第一次运行, 需重新生成 poll queue
		// 检查是否存在失效轮训池
		s.clearExpirePoll()

		// 生成新分配池
		var idstrings []string
		idstrings, err = canAssignRoles(s.TicketItem)
		if err != nil {
			return
		}
		adminUids, _, _ := models.GetUserIDsByRoleIDStringsFromDB(idstrings)
		if len(adminUids) <= 0 {
			err = fmt.Errorf("[pollAssign] there is no users on the ticket item config role, ticket item:%d, role: %s",
				s.TicketItem, idstrings)
			logs.Error(err)
			return
		}
		for _, v := range adminUids {
			redisCli.Do("LPUSH", key, v)
			adminUID = v
		}
		return
	}

	if v, ok := out.([]byte); ok {
		adminUID, _ = strconv.ParseInt(string(v), 10, 64)
		return
	}
	err = fmt.Errorf("[pollAssign] pop user is not vaild int64, ticket item:%d, poll_queue: %s, pop user: %s",
		s.TicketItem, key, out)
	logs.Error(err)
	return
}

// 获取分配池名称(redis queue name)
func (s *PollingWorkerStrategy) getAssignPollKey() (key string, err error) {
	keyPrefix, err := s.getAssignPollKeyPrefix()
	if err != nil {
		return
	}
	idstrings, err := canAssignRoles(s.TicketItem)
	if err != nil {
		// err was already logged
		return
	}

	itemKeyPrefix := keyPrefix + createAssignQueueNameItemPart(s.TicketItem)
	key = itemKeyPrefix + createAssignQueueNameRolePart(idstrings...)
	return
}

// 获取分配池名称前缀(redis queue name prefix)
func (s *PollingWorkerStrategy) getAssignPollKeyPrefix() (prefix string, err error) {
	prefix = beego.AppConfig.String("queue_ticket_assign_poll_prefix")
	if len(prefix) <= 0 {
		err = fmt.Errorf("[getAssignPollKeyPrefix] queue_ticket_assign_poll_prefix must be configured in redis-key.conf for ticket item")
		//logs.Error(err)
	}
	return
}

func (s *PollingWorkerStrategy) getMatchAssignQueuePollByTicketItem() (string, error) {
	keyPrefix, err := s.getAssignPollKeyPrefix()
	if err != nil {
		return "", err
	}
	if len(keyPrefix) <= 5 {
		return "", fmt.Errorf("[getMatchAssignQueuePollByTicketItem]Fuzzy matching key rule is too short , not secure, key rule:%s", keyPrefix)
	}
	itemKeyPrefix := keyPrefix + createAssignQueueNameItemPart(s.TicketItem) + "*"
	return itemKeyPrefix, nil
}

func (s *PollingWorkerStrategy) getMatchRuleAssignQueuePollByRole(roleID int64) (string, error) {
	keyPrefix, err := s.getAssignPollKeyPrefix()
	if err != nil {
		return "", err
	}
	if len(keyPrefix) <= 5 {
		return "", fmt.Errorf("[getMatchAssignQueuePollByTicketItem]Fuzzy matching key prefix  is too short , not secure, key rule:%s", keyPrefix)
	}
	itemKeyPrefix := keyPrefix + "*" + createAssignQueueNameRolePart(strconv.FormatInt(roleID, 10)) + "*"
	return itemKeyPrefix, nil
}

func (s *PollingWorkerStrategy) clearExpirePoll() {
	macthKey, err := s.getMatchAssignQueuePollByTicketItem()
	if err != nil {
		logs.Error(err)
		return
	}
	redisCli := storage.RedisStorageClient.Get()
	defer redisCli.Close()
	redisKeys, _ := redisCli.Do("KEYS", macthKey)
	if expiredQueue, ok := redisKeys.([]interface{}); ok && len(expiredQueue) > 0 {
		for _, v := range expiredQueue {
			if queueName, ok := v.([]byte); ok {
				redisCli.Do("DEL", string(queueName))
			}
			// 危险操作
		}
	}
}

// DayAvgAlternateWorkerStrategy 日均分工单, 人力候补策略
// 日定额工单, 未分完工单的, 人力策略, 暂定随机
type DayAvgAlternateWorkerStrategy struct {
	TicketItem   types.TicketItemEnum
	activeWorker []int64
	next         int
	isInit       bool
}

// OneWorker 此处为近实时数据, 不依赖cache
func (s *DayAvgAlternateWorkerStrategy) OneWorker() (adminUID int64, err error) {
	if !s.isInit {
		s.init()
	}
	activeNum := len(s.activeWorker)
	if activeNum == 0 {
		err = fmt.Errorf("[dayFinalAssign] no active user for assign , will quit, activeNum:%d, ;TicketItem: %d", activeNum, s.TicketItem)
		return
	}
	logs.Debug("[DayAvgAlternateWorkerStrategy] OneWorker , worker index:", s.next)
	adminUID = s.activeWorker[s.next]
	logs.Debug("[DayAvgAlternateWorkerStrategy] OneWorker , worker uid:", adminUID)
	s.cycleQueueNext()
	return
}

func (s *DayAvgAlternateWorkerStrategy) cycleQueueNext() {
	if len(s.activeWorker) == s.next+1 {
		s.next = 0
		return
	}
	s.next++
}

func (s *DayAvgAlternateWorkerStrategy) init() {
	admins := activeCanAssignUsersByTicketItem(s.TicketItem)
	for _, admin := range admins {
		s.activeWorker = append(s.activeWorker, admin.Id)
	}
	logs.Debug("[DayAvgAlternateWorkerStrategy] first get queue:", s.activeWorker)
	rand.Shuffle(len(s.activeWorker), func(i int, j int) {
		s.activeWorker[i], s.activeWorker[j] = s.activeWorker[j], s.activeWorker[i]
	})
	logs.Debug("[DayAvgAlternateWorkerStrategy] shuffle queue:", s.activeWorker)
	s.next = 0
	s.isInit = true
}

// UserStartWork no cache, and not persistent job , ignore
func (s *DayAvgAlternateWorkerStrategy) UserStartWork(adminUID, roleID int64) {
	return
}

// UserStopWork no cache, and not persistent job , ignore
func (s *DayAvgAlternateWorkerStrategy) UserStopWork(adminUID, roleID int64) {
	return
}

// EmptyWorkerStrategy 空策略， 仅仅是为了减少代码处理逻辑
type EmptyWorkerStrategy struct {
	TicketItem types.TicketItemEnum
}

// OneWorker 此处为近实时数据, 不依赖cache
func (s *EmptyWorkerStrategy) OneWorker() (adminUID int64, err error) {
	return
}

// UserStartWork no cache, and not persistent job , ignore
func (s *EmptyWorkerStrategy) UserStartWork(adminUID, roleID int64) {
	return
}

// UserStopWork no cache, and not persistent job , ignore
func (s *EmptyWorkerStrategy) UserStopWork(adminUID, roleID int64) {
	return
}

//

// 公用辅助方法

func createAssignQueueNameRolePart(idstrings ...string) (partName string) {
	for _, idstring := range idstrings {
		partName += fmt.Sprintf(types.TicketQueueNameRoleIDVar, idstring)
	}
	return
}

func createAssignQueueNameItemPart(ticketItem types.TicketItemEnum) (partName string) {
	partName += fmt.Sprintf(string(types.TicketQueueNameItemVar), ticketItem)
	return
}

package main

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/pkg/schema_task"

	_ "github.com/astaxie/beego/session/redis"

	"github.com/astaxie/beego/logs"

	"micro-loan/common/lib/clogs"
	"micro-loan/common/lib/redis/pubsub"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego"
	"github.com/gomodule/redigo/redis"
)

type TaskSchemaInfo struct {
	Id         int64
	StartDate  int64
	EndDate    int64
	RunTime    []int
	TimeIndex  int
	NextTime   int64
	SchemaMode types.SchemaMode
	IsStart    bool
	FuncName   string
}

func (c *TaskSchemaInfo) SetNext() {
	base := tools.NaturalDay(0)
	if base < c.StartDate {
		base = c.StartDate
	}

	if c.NextTime == 0 {
		c.NextTime = base + int64(c.RunTime[0])

		return
	}

	if c.NextTime > tools.GetUnixMillis()/1000*1000 {
		return
	}

	if c.TimeIndex < len(c.RunTime)-1 {
		c.NextTime = c.NextTime + int64(c.RunTime[c.TimeIndex+1]-c.RunTime[c.TimeIndex])
		c.TimeIndex++
	} else {
		c.NextTime = base + tools.MILLSSECONDADAY + int64(c.RunTime[0])
		c.TimeIndex = 0
	}
}

type TaskList []*TaskSchemaInfo

func (p TaskList) Len() int { return len(p) }

func (p TaskList) Less(i, j int) bool {
	return p[i].NextTime < p[j].NextTime
}

func (p TaskList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

var stop chan bool
var changed chan models.SchemaInfo
var taskList TaskList
var taskMap map[int64]*TaskSchemaInfo

func init() {
	taskList = make(TaskList, 0)
	taskMap = make(map[int64]*TaskSchemaInfo)
	changed = make(chan models.SchemaInfo)
	stop = make(chan bool)
}

func handleUpdateEvent() {
	c := make(chan []byte)
	schemaInfo := models.SchemaInfo{}
	go pubsub.Subscribe(schemaInfo.TableName(), c)

	for {
		select {
		case d := <-c:
			m := models.SchemaInfo{}
			json.Unmarshal(d, &m)
			logs.Info("[handleUpdateEvent] receive data ch:%s, data:%v", m.TableName(), m)
			changed <- m
			logs.Debug("[handleUpdateEvent] push data to chan ch:%s, data:%v", m.TableName(), m)
		}
	}
}

func handleStopEvent() {
	qName := beego.AppConfig.String("schema")

	for {
		storageClient := storage.RedisStorageClient.Get()

		id, _ := redis.Int64(storageClient.Do("RPOP", qName))

		storageClient.Close()

		if id == types.TaskExitCmd {
			logs.Info("[handleStopEvent] receive stop event now:%d", tools.GetUnixMillis())
			stop <- true
			return
		}

		time.Sleep(time.Second)
	}
}

func Start() {
	list, err := models.LoadSchemaInfo()
	if err != nil {
		logs.Error("[Start] LoadSchemaInfo error err:%v", err)
	} else {
		logs.Debug("[Start] LoadSchemaInfo success size:%d", len(list))
	}

	for _, v := range list {
		updateSchema(v)
	}

	sort.Sort(taskList)

	go handleUpdateEvent()

	index := 0
	for {
		if len(taskList) == 0 {
			logs.Info("[Start] taskList empty wait for changed")
			m := <-changed
			updateSchema(m)
			sort.Sort(taskList)
			index = 0
		}

		if index >= len(taskList) {
			index = 0
		}

		now := tools.GetUnixMillis() / 1000 * 1000

		curTask := taskList[index]

		if !curTask.IsStart {
			taskList = append(taskList[:index], taskList[index+1:]...)
			index = 0
			continue
		}

		if now > curTask.EndDate {
			service.StopSchema(curTask.Id)

			taskList = append(taskList[:index], taskList[index+1:]...)
			index = 0
			continue
		}

		if curTask.NextTime <= now {
			if now-curTask.NextTime < 60*1000 {
				logs.Info("[Start] RunSchema taskId:%d, runTime:%d", curTask.Id, curTask.NextTime)
				go service.RunSchema(curTask.Id, curTask.NextTime)
			}

			curTask.SetNext()
			sort.Sort(taskList)
			index = 0
			continue
		}

		d := curTask.NextTime - now
		timer := time.NewTimer(time.Millisecond * time.Duration(d))

		logs.Info("[Start] taskId:%d, nextTime:%d, now:%d, sleep:%d", curTask.Id, curTask.NextTime, now, d)

		select {
		case m := <-changed:
			logs.Info("[Start] receive changed event data:%v", m)
			timer.Stop()
			updateSchema(m)
			sort.Sort(taskList)
			index = 0
		case <-timer.C:
			logs.Debug("[Start] timer return sleep:%d", d)
			continue
		case <-stop:
			logs.Warn("[Start] break cycle")
			return
		}
	}
}

func updateSchema(info models.SchemaInfo) {
	if v, ok := taskMap[info.Id]; !ok {
		if info.SchemaStatus != types.SchemaStatusOff {
			times := parseTime(info.SchemaTime)
			if len(times) == 0 {
				return
			}

			task := new(TaskSchemaInfo)
			task.Id = info.Id
			task.StartDate = info.StartDate
			task.EndDate = info.EndDate
			task.RunTime = times
			task.NextTime = 0
			task.TimeIndex = 0
			task.SchemaMode = info.SchemaMode
			task.IsStart = true
			task.FuncName = info.FuncName
			task.SetNext()

			taskList = append(taskList, task)
			taskMap[info.Id] = task
		}
	} else {
		if info.SchemaStatus != types.SchemaStatusOff {
			task := v

			times := parseTime(info.SchemaTime)
			if len(times) == 0 {
				task.IsStart = false
				return
			}

			task.Id = info.Id
			task.StartDate = info.StartDate
			task.EndDate = info.EndDate
			task.RunTime = times
			task.NextTime = 0
			task.TimeIndex = 0
			task.SchemaMode = info.SchemaMode
			task.IsStart = true
			task.FuncName = info.FuncName
			task.SetNext()
		} else {
			task := v
			task.IsStart = false
		}
	}
}

func parseTime(str string) []int {
	ret := make([]int, 0)

	list := strings.Split(str, ",")
	for _, v := range list {
		times := strings.Split(v, ":")
		if len(times) < 2 {
			continue
		}
		h, _ := tools.Str2Int(strings.Trim(times[0], " "))
		m, _ := tools.Str2Int(strings.Trim(times[1], " "))

		h -= 7
		if h < 0 {
			h += 24
		}

		t := (h*60 + m) * 60 * 1000
		ret = append(ret, t)
	}

	sort.Ints(ret)
	return ret
}

func main() {
	dir := beego.AppConfig.String("log_dir")
	clogs.InitLog(dir, "schema")

	go schema_task.StartPushBackup()
	go schema_task.StartSmsBackup()
	go schema_task.StartCouponBackup()

	go handleStopEvent()

	Start()

	close(changed)
	close(stop)
}

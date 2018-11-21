package main

import (
	"fmt"
	"io/ioutil"

	// 数据库初始化
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/thirdparty/xendit"
	"micro-loan/common/tools"

	"encoding/json"

	"micro-loan/common/models"

	"io"
	"os"

	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"
)

var qs = ""

type Quest struct {
	ExternalId       string `json:"external_id"`
	OldId            string `json:"old_id"`
	OldAccountNumber string `json:"old_account_number"`
	NewId            string `json:"new_id"`
	NewAccountNumber string `json:"new_account_number"`
}

func fixQuest(quest Quest) {
	eA, err := models.GetEAccountByENumber(quest.OldAccountNumber)
	if err != nil {
		logs.Error("[fixQuest] err:%v quest:%#v", err, quest)
		return
	}

	var resp = xendit.XenditCreateVAccountResponse{}
	err = json.Unmarshal([]byte(eA.CallbackJson), &resp)
	if err != nil {
		logs.Error("[Unmarshal] err:%v quest:%#v eA:%#v", err, quest, eA)
		return
	}

	if tools.Int642Str(eA.UserAccountId) != quest.ExternalId {
		logs.Error("UserAccountId not equal .  quest:%#v eA:%#v", quest, eA)
		saveQuest(quest)
		return
	}

	old := eA
	eA.EAccountNumber = quest.NewAccountNumber
	if eA.Status != "ACTIVE" {
		eA.Status = "ACTIVE"
	}
	resp.AccountNumber = quest.NewAccountNumber
	cj, _ := json.Marshal(resp)
	eA.CallbackJson = string(cj)

	_, err = eA.UpdateEAccount(&eA)
	if err != nil {
		logs.Error("[OrmUpdate] err:%v quest:%#v eA:%#v", err, quest, eA)
		saveQuest(quest)
		return
	}
	models.OpLogWrite(0, eA.UserAccountId, models.OpCodeAccountBaseUpdate, old.TableName(), old, eA)
}

func saveQuest(quest Quest) {
	s, _ := json.Marshal(quest)
	WriteFile("err.json", []byte(s), 0644)
	WriteFile("err.json", []byte("\n"), 0644)
}

func main() {

	// 设置进程 title
	// +1 分布式锁
	// -1 正常退出时,释放锁
	procTitle := "fix_xendit_va"
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
	defer storageClient.Do("DEL", lockKey)

	qL := []Quest{}
	err = json.Unmarshal([]byte(qs), &qL)
	if err != nil || len(qL) == 0 {
		logs.Error("Unmarshal err:%v qs:%s", err, qs)
		return
	}

	logs.Info("len:%d", len(qL))
	for k, v := range qL {
		logs.Info("start fix[%d] quest:%#v", k, v)
		fixQuest(v)
	}

	logs.Notice("ok")
}

func init() {
	b, err := ioutil.ReadFile("va_pair.json")
	if err != nil {
		logs.Error("file Open err:%v", err)
		panic(1)
	}
	qs = string(b)
}

func WriteFile(filename string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, perm)
	if err != nil {
		return err
	}
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

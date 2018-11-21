package models

import (
	"github.com/astaxie/beego/logs"

	"micro-loan/common/tools"
	"micro-loan/common/types"
)

const LOG_ACCOUNT_BASE_EXT_TABLENAME string = "log_account_base_ext"

type LogAccountBaseExt struct {
	Id                           int64 `orm:"pk"`
	AccountId                    int64
	AuthorizeStatusYys           int
	AuthorizeFinishTimeYys       int64
	AuthorizeCrawleTimeYys       int64
	QuotaRaiseYys                int64
	AuthorizeStatusGoJek         int
	AuthorizeFinishTimeGoJek     int64
	AuthorizeCrawleTimeGoJek     int64
	QuotaRaiseGoJek              int64
	AuthorizeStatusLazada        int
	AuthorizeFinishTimeLazada    int64
	AuthorizeCrawleTimeLazada    int64
	QuotaRaiseLazada             int64
	AuthorizeStatusTokopedia     int
	AuthorizeFinishTimeTokopedia int64
	AuthorizeCrawleTimeTokopedia int64
	QuotaRaiseTokopedia          int64
	AuthorizeStatusFacebook      int
	AuthorizeFinishTimeFacebook  int64
	AuthorizeCrawleTimeFacebook  int64
	QuotaRaiseFacebook           int64
	AuthorizeStatusInstagram     int
	AuthorizeFinishTimeInstagram int64
	AuthorizeCrawleTimeInstagram int64
	QuotaRaiseInstagram          int64
	AuthorizeStatusLinkedin      int
	AuthorizeFinishTimeLinkedin  int64
	AuthorizeCrawleTimeLinkedin  int64
	QuotaRaiseLinkedin           int64
	RecallTag                    int
	NpwpNo                       string
	NpwpStatus                   int
	NpwpTime                     int64
	QuotaRaiseNpwp               int64
	PushMessageMark              int64
	PhyInvalidTag                int
	Ctime                        int64
	Utime                        int64
	CtimeRecord                  int64
}

func (r *LogAccountBaseExt) TableName() string {
	return LOG_ACCOUNT_BASE_EXT_TABLENAME
}

func (r *LogAccountBaseExt) Using() string {
	return types.OrmDataBaseApi
}
func (r *LogAccountBaseExt) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *LogAccountBaseExt) InsertWithNoReturn() error {
	// 新增
	_, err := OrmInsert(r)
	if err != nil {
		logs.Error("[InsertWithNoReturn] OrmInsert err:%v, r:%#v", err, r)
	}

	return err
}

func InsertLogAccountBaseExt(aExt AccountBaseExt) error {
	laExt := LogAccountBaseExt{
		AccountId:                    aExt.AccountId,
		AuthorizeStatusYys:           aExt.AuthorizeStatusYys,
		AuthorizeFinishTimeYys:       aExt.AuthorizeFinishTimeYys,
		AuthorizeCrawleTimeYys:       aExt.AuthorizeCrawleTimeYys,
		QuotaRaiseYys:                aExt.QuotaRaiseYys,
		AuthorizeStatusGoJek:         aExt.AuthorizeStatusGoJek,
		AuthorizeFinishTimeGoJek:     aExt.AuthorizeFinishTimeGoJek,
		AuthorizeCrawleTimeGoJek:     aExt.AuthorizeCrawleTimeGoJek,
		QuotaRaiseGoJek:              aExt.QuotaRaiseGoJek,
		AuthorizeStatusLazada:        aExt.AuthorizeStatusLazada,
		AuthorizeFinishTimeLazada:    aExt.AuthorizeFinishTimeLazada,
		AuthorizeCrawleTimeLazada:    aExt.AuthorizeCrawleTimeLazada,
		QuotaRaiseLazada:             aExt.QuotaRaiseLazada,
		AuthorizeStatusTokopedia:     aExt.AuthorizeStatusTokopedia,
		AuthorizeFinishTimeTokopedia: aExt.AuthorizeFinishTimeTokopedia,
		AuthorizeCrawleTimeTokopedia: aExt.AuthorizeCrawleTimeTokopedia,
		QuotaRaiseTokopedia:          aExt.QuotaRaiseTokopedia,
		AuthorizeStatusFacebook:      aExt.AuthorizeStatusFacebook,
		AuthorizeFinishTimeFacebook:  aExt.AuthorizeFinishTimeFacebook,
		AuthorizeCrawleTimeFacebook:  aExt.AuthorizeCrawleTimeFacebook,
		QuotaRaiseFacebook:           aExt.QuotaRaiseFacebook,
		AuthorizeStatusInstagram:     aExt.AuthorizeStatusInstagram,
		AuthorizeFinishTimeInstagram: aExt.AuthorizeFinishTimeInstagram,
		AuthorizeCrawleTimeInstagram: aExt.AuthorizeCrawleTimeInstagram,
		QuotaRaiseInstagram:          aExt.QuotaRaiseInstagram,
		AuthorizeStatusLinkedin:      aExt.AuthorizeStatusLinkedin,
		AuthorizeFinishTimeLinkedin:  aExt.AuthorizeFinishTimeLinkedin,
		AuthorizeCrawleTimeLinkedin:  aExt.AuthorizeCrawleTimeLinkedin,
		QuotaRaiseLinkedin:           aExt.QuotaRaiseLinkedin,
		RecallTag:                    aExt.RecallTag,
		NpwpNo:                       aExt.NpwpNo,
		NpwpStatus:                   aExt.NpwpStatus,
		NpwpTime:                     aExt.NpwpTime,
		QuotaRaiseNpwp:               aExt.QuotaRaiseNpwp,
		PushMessageMark:              aExt.PushMessageMark,
		PhyInvalidTag:                aExt.PhyInvalidTag,
		Ctime:                        aExt.Ctime,
		Utime:                        aExt.Utime,
		CtimeRecord:                  tools.GetUnixMillis(),
	}
	return laExt.InsertWithNoReturn()
}

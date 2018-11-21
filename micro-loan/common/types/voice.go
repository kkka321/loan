package types

type VoiceType int

const (
	// 还款提醒
	VoiceTypeYesterday VoiceType = -1 // 昨天到期，逾期一天
	VoiceTypeToday     VoiceType = 0  // 今天到期
	VoiceTypeTomorrow  VoiceType = 1  // 明天到期
	VoiceTypeOverdue   VoiceType = 11 // 逾期案件自动外呼标记

	// InfoReview工单自动呼叫
	VoiceTypeInfoReview VoiceType = 50 // InfoReview自动呼叫语音类型
)

const (
	VoiceCallSuccess = 1 // 语音群呼API成功
)

const (
	CallTimeDelimiter = "," // 呼叫时间配置间隔
	CallTimeBlankVal  = "-" // 取消呼叫时，配置参数为'-'
)

// InfoReview工单自动外呼定时器key前缀
const (
	InfoReviewAutoCallKeyPre = "info_review_auto_call"
)

// InfoReview工单自动外呼配置
const (
	InfoReviewAutoCallNumName        = "inforeview_call_num"
	InfoReviewAutoCallTimeName       = "inforeview_call_time"
	InfoReviewAutoCallRecordFileName = "inforeview_call_record_file"
)

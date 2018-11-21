package service

import "micro-loan/common/pkg/system/config"

func GetRepeatLoanQuota() (on bool) {
	on, _ = config.ValidItemBool("abtest_repeat_loan_quota")

	return
}

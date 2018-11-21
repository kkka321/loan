package product

import (
	// _ "micro-loan/common/lib/clogs"
	// _ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/models"
	"micro-loan/common/types"
	"testing"

	"github.com/astaxie/beego/logs"
)

//TrialCalcRepayTypeOnce(trialIn types.ProductTrialCalcIn, product models.Product) (trialResults []types.ProductTrialCalcResult, err error)
func TestTrialCalcRepayTypeOnce(t *testing.T) {
	test := []struct {
		trialIn      types.ProductTrialCalcIn
		trialResults []types.ProductTrialCalcResult
		err          error
	}{
		// 等待还款
		{types.ProductTrialCalcIn{
			ID:           180508030000776897,
			Loan:         0,
			Amount:       5000,
			Period:       7,
			LoanDate:     "2018-05-01",
			CurrentDate:  "2018-05-05",
			RepayDate:    "0",
			RepayedTotal: 0,
		}, []types.ProductTrialCalcResult{
			// 0期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          0,
				RepayDateShould:          1525104000000,
				Loan:                     4860,
				OverdueDays:              0,
				RepayStatus:              0,
				RepayTotalShould:         140,
				RepayAmountShould:        0,
				RepayInterestShould:      0,
				RepayFeeShould:           140,
				RepayGraceInterestShould: 0,
				RepayPenaltyShould:       0,
				ForfeitPenalty:           0,
				RepayedDate:              1525104000000,
				RepayedTotal:             140,
				RepayedAmount:            0,
				RepayedInterest:          0,
				RepayedGraceInterest:     0,
				RepayedFee:               140,
				RepayedPenalty:           0,
				RepayedForfeitPenalty:    0,
			},
			// 1期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          1,
				RepayDateShould:          1525708800000,
				Loan:                     0,
				OverdueDays:              0,
				RepayStatus:              int(types.LoanStatusWaitRepayment),
				RepayTotalShould:         5350,
				RepayAmountShould:        5000,
				RepayInterestShould:      350,
				RepayFeeShould:           0,
				RepayGraceInterestShould: 0,
				RepayPenaltyShould:       0,
				ForfeitPenalty:           0,
				RepayedDate:              0,
				RepayedTotal:             0,
				RepayedAmount:            0,
				RepayedInterest:          0,
				RepayedGraceInterest:     0,
				RepayedFee:               0,
				RepayedPenalty:           0,
				RepayedForfeitPenalty:    0,
			},
		},
			nil},

		// 还款期内 部分还款
		{types.ProductTrialCalcIn{
			ID:           180508030000776897,
			Loan:         0,
			Amount:       5000,
			Period:       7,
			LoanDate:     "2018-05-01",
			CurrentDate:  "2018-05-08",
			RepayDate:    "2018-05-08",
			RepayedTotal: 5000,
		}, []types.ProductTrialCalcResult{
			// 0期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          0,
				RepayDateShould:          1525104000000,
				Loan:                     4860,
				OverdueDays:              0,
				RepayStatus:              0,
				RepayTotalShould:         140,
				RepayAmountShould:        0,
				RepayInterestShould:      0,
				RepayFeeShould:           140,
				RepayGraceInterestShould: 0,
				RepayPenaltyShould:       0,
				ForfeitPenalty:           0,
				RepayedDate:              1525104000000,
				RepayedTotal:             140,
				RepayedAmount:            0,
				RepayedInterest:          0,
				RepayedGraceInterest:     0,
				RepayedFee:               140,
				RepayedPenalty:           0,
				RepayedForfeitPenalty:    0,
			},
			// 1期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          1,
				RepayDateShould:          1525708800000,
				Loan:                     0,
				OverdueDays:              0,
				RepayStatus:              int(types.LoanStatusPartialRepayment),
				RepayTotalShould:         5350,
				RepayAmountShould:        5000,
				RepayInterestShould:      350,
				RepayFeeShould:           0,
				RepayGraceInterestShould: 0,
				RepayPenaltyShould:       0,
				ForfeitPenalty:           0,
				RepayedDate:              1525708800000,
				RepayedTotal:             5000,
				RepayedAmount:            4650,
				RepayedInterest:          350,
				RepayedGraceInterest:     0,
				RepayedFee:               0,
				RepayedPenalty:           0,
				RepayedForfeitPenalty:    0,
			},
		},
			nil},

		// 还款期内 还清
		{types.ProductTrialCalcIn{
			ID:           180508030000776897,
			Loan:         0,
			Amount:       5000,
			Period:       7,
			LoanDate:     "2018-05-01",
			CurrentDate:  "2018-05-08",
			RepayDate:    "2018-05-08",
			RepayedTotal: 6000,
		}, []types.ProductTrialCalcResult{
			// 0期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          0,
				RepayDateShould:          1525104000000,
				Loan:                     4860,
				OverdueDays:              0,
				RepayStatus:              0,
				RepayTotalShould:         140,
				RepayAmountShould:        0,
				RepayInterestShould:      0,
				RepayFeeShould:           140,
				RepayGraceInterestShould: 0,
				RepayPenaltyShould:       0,
				ForfeitPenalty:           0,
				RepayedDate:              1525104000000,
				RepayedTotal:             140,
				RepayedAmount:            0,
				RepayedInterest:          0,
				RepayedGraceInterest:     0,
				RepayedFee:               140,
				RepayedPenalty:           0,
				RepayedForfeitPenalty:    0,
			},
			// 1期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          1,
				RepayDateShould:          1525708800000,
				Loan:                     0,
				OverdueDays:              0,
				RepayStatus:              int(types.LoanStatusAlreadyCleared),
				RepayTotalShould:         5350,
				RepayAmountShould:        5000,
				RepayInterestShould:      350,
				RepayFeeShould:           0,
				RepayGraceInterestShould: 0,
				RepayPenaltyShould:       0,
				ForfeitPenalty:           0,
				RepayedDate:              1525708800000,
				RepayedTotal:             5350,
				RepayedAmount:            5000,
				RepayedInterest:          350,
				RepayedGraceInterest:     0,
				RepayedFee:               0,
				RepayedPenalty:           0,
				RepayedForfeitPenalty:    0,
			},
		},
			nil},

		// 还款期外 未还款 未还清
		{types.ProductTrialCalcIn{
			ID:           180508030000776897,
			Loan:         0,
			Amount:       5000,
			Period:       7,
			LoanDate:     "2018-05-01",
			CurrentDate:  "2018-05-18",
			RepayDate:    "",
			RepayedTotal: 0,
		}, []types.ProductTrialCalcResult{
			// 0期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          0,
				RepayDateShould:          1525104000000,
				Loan:                     4860,
				OverdueDays:              0,
				RepayStatus:              0,
				RepayTotalShould:         140,
				RepayAmountShould:        0,
				RepayInterestShould:      0,
				RepayFeeShould:           140,
				RepayGraceInterestShould: 0,
				RepayPenaltyShould:       0,
				ForfeitPenalty:           0,
				RepayedDate:              1525104000000,
				RepayedTotal:             140,
				RepayedAmount:            0,
				RepayedInterest:          0,
				RepayedGraceInterest:     0,
				RepayedFee:               140,
				RepayedPenalty:           0,
				RepayedForfeitPenalty:    0,
			},
			// 1期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          1,
				RepayDateShould:          1525708800000,
				Loan:                     0,
				OverdueDays:              10,
				RepayStatus:              int(types.LoanStatusOverdue),
				RepayTotalShould:         6300,
				RepayAmountShould:        5000,
				RepayInterestShould:      350,
				RepayFeeShould:           0,
				RepayGraceInterestShould: 50,
				RepayPenaltyShould:       900,
				ForfeitPenalty:           0,
				RepayedDate:              0,
				RepayedTotal:             0,
				RepayedAmount:            0,
				RepayedInterest:          0,
				RepayedGraceInterest:     0,
				RepayedFee:               0,
				RepayedPenalty:           0,
				RepayedForfeitPenalty:    0,
			},
		},
			nil},

		// 还款期外 部分还款 未还清
		{types.ProductTrialCalcIn{
			ID:           180508030000776897,
			Loan:         0,
			Amount:       5000,
			Period:       7,
			LoanDate:     "2018-05-01",
			CurrentDate:  "2018-05-18",
			RepayDate:    "2018-05-12",
			RepayedTotal: 5000,
		}, []types.ProductTrialCalcResult{
			// 0期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          0,
				RepayDateShould:          1525104000000,
				Loan:                     4860,
				OverdueDays:              0,
				RepayStatus:              0,
				RepayTotalShould:         140,
				RepayAmountShould:        0,
				RepayInterestShould:      0,
				RepayFeeShould:           140,
				RepayGraceInterestShould: 0,
				RepayPenaltyShould:       0,
				ForfeitPenalty:           0,
				RepayedDate:              1525104000000,
				RepayedTotal:             140,
				RepayedAmount:            0,
				RepayedInterest:          0,
				RepayedGraceInterest:     0,
				RepayedFee:               140,
				RepayedPenalty:           0,
				RepayedForfeitPenalty:    0,
			},
			// 1期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          1,
				RepayDateShould:          1525708800000,
				Loan:                     0,
				OverdueDays:              10,
				RepayStatus:              int(types.LoanStatusOverdue),
				RepayTotalShould:         5784,
				RepayAmountShould:        5000,
				RepayInterestShould:      350,
				RepayFeeShould:           0,
				RepayGraceInterestShould: 50,
				RepayPenaltyShould:       384,
				ForfeitPenalty:           0,
				RepayedDate:              1526054400000,
				RepayedTotal:             5000,
				RepayedAmount:            4300,
				RepayedInterest:          350,
				RepayedGraceInterest:     50,
				RepayedFee:               0,
				RepayedPenalty:           300,
				RepayedForfeitPenalty:    0,
			},
		},
			nil},

		// 还款期外 还清
		{types.ProductTrialCalcIn{
			ID:           180508030000776897,
			Loan:         0,
			Amount:       5000,
			Period:       7,
			LoanDate:     "2018-05-01",
			CurrentDate:  "2018-05-18",
			RepayDate:    "2018-05-12",
			RepayedTotal: 6000,
		}, []types.ProductTrialCalcResult{
			// 0期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          0,
				RepayDateShould:          1525104000000,
				Loan:                     4860,
				OverdueDays:              0,
				RepayStatus:              0,
				RepayTotalShould:         140,
				RepayAmountShould:        0,
				RepayInterestShould:      0,
				RepayFeeShould:           140,
				RepayGraceInterestShould: 0,
				RepayPenaltyShould:       0,
				ForfeitPenalty:           0,
				RepayedDate:              1525104000000,
				RepayedTotal:             140,
				RepayedAmount:            0,
				RepayedInterest:          0,
				RepayedGraceInterest:     0,
				RepayedFee:               140,
				RepayedPenalty:           0,
				RepayedForfeitPenalty:    0,
			},
			// 1期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          1,
				RepayDateShould:          1525708800000,
				Loan:                     0,
				OverdueDays:              0,
				RepayStatus:              int(types.LoanStatusAlreadyCleared),
				RepayTotalShould:         5700,
				RepayAmountShould:        5000,
				RepayInterestShould:      350,
				RepayFeeShould:           0,
				RepayGraceInterestShould: 50,
				RepayPenaltyShould:       300,
				ForfeitPenalty:           0,
				RepayedDate:              1526054400000,
				RepayedTotal:             5700,
				RepayedAmount:            5000,
				RepayedInterest:          350,
				RepayedGraceInterest:     50,
				RepayedFee:               0,
				RepayedPenalty:           300,
				RepayedForfeitPenalty:    0,
			},
		},
			nil},

		// 还款期外大于90天 未还款 未还清
		{types.ProductTrialCalcIn{
			ID:           180508030000776897,
			Loan:         0,
			Amount:       5000,
			Period:       7,
			LoanDate:     "2018-06-01",
			CurrentDate:  "2018-09-10",
			RepayDate:    "0",
			RepayedTotal: 0,
		}, []types.ProductTrialCalcResult{
			// 0期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          0,
				RepayDateShould:          1527782400000,
				Loan:                     4860,
				OverdueDays:              0,
				RepayStatus:              0,
				RepayTotalShould:         140,
				RepayAmountShould:        0,
				RepayInterestShould:      0,
				RepayFeeShould:           140,
				RepayGraceInterestShould: 0,
				RepayPenaltyShould:       0,
				ForfeitPenalty:           0,
				RepayedDate:              1527782400000,
				RepayedTotal:             140,
				RepayedAmount:            0,
				RepayedInterest:          0,
				RepayedGraceInterest:     0,
				RepayedFee:               140,
				RepayedPenalty:           0,
				RepayedForfeitPenalty:    0,
			},
			// 1期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          1,
				RepayDateShould:          1528387200000,
				Loan:                     0,
				OverdueDays:              94,
				RepayStatus:              int(types.LoanStatusOverdue),
				RepayTotalShould:         14300,
				RepayAmountShould:        5000,
				RepayInterestShould:      350,
				RepayFeeShould:           0,
				RepayGraceInterestShould: 50,
				RepayPenaltyShould:       8900,
				ForfeitPenalty:           0,
				RepayedDate:              0,
				RepayedTotal:             0,
				RepayedAmount:            0,
				RepayedInterest:          0,
				RepayedGraceInterest:     0,
				RepayedFee:               0,
				RepayedPenalty:           0,
				RepayedForfeitPenalty:    0,
			},
		},
			nil},

		// 当前时间还款期外大于90天 部分还款在90天内 未还清
		{types.ProductTrialCalcIn{
			ID:           180508030000776897,
			Loan:         0,
			Amount:       5000,
			Period:       7,
			LoanDate:     "2018-06-01",
			CurrentDate:  "2018-09-10",
			RepayDate:    "2018-09-01",
			RepayedTotal: 8000,
		}, []types.ProductTrialCalcResult{
			// 0期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          0,
				RepayDateShould:          1527782400000,
				Loan:                     4860,
				OverdueDays:              0,
				RepayStatus:              0,
				RepayTotalShould:         140,
				RepayAmountShould:        0,
				RepayInterestShould:      0,
				RepayFeeShould:           140,
				RepayGraceInterestShould: 0,
				RepayPenaltyShould:       0,
				ForfeitPenalty:           0,
				RepayedDate:              1527782400000,
				RepayedTotal:             140,
				RepayedAmount:            0,
				RepayedInterest:          0,
				RepayedGraceInterest:     0,
				RepayedFee:               140,
				RepayedPenalty:           0,
				RepayedForfeitPenalty:    0,
			},
			// 1期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          1,
				RepayDateShould:          1528387200000,
				Loan:                     0,
				OverdueDays:              94,
				RepayStatus:              int(types.LoanStatusOverdue),
				RepayTotalShould:         14300,
				RepayAmountShould:        5000,
				RepayInterestShould:      350,
				RepayFeeShould:           0,
				RepayGraceInterestShould: 50,
				RepayPenaltyShould:       8900,
				ForfeitPenalty:           0,
				RepayedDate:              1535731200000,
				RepayedTotal:             8000,
				RepayedAmount:            0,
				RepayedInterest:          0,
				RepayedGraceInterest:     0,
				RepayedFee:               0,
				RepayedPenalty:           8000,
				RepayedForfeitPenalty:    0,
			},
		},
			nil},

		// 当前时间还款期外大于90天 部分还款在90天内 未还清
		{types.ProductTrialCalcIn{
			ID:           180508030000776897,
			Loan:         0,
			Amount:       5000,
			Period:       7,
			LoanDate:     "2018-06-01",
			CurrentDate:  "2018-09-10",
			RepayDate:    "2018-09-01",
			RepayedTotal: 9000,
		}, []types.ProductTrialCalcResult{
			// 0期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          0,
				RepayDateShould:          1527782400000,
				Loan:                     4860,
				OverdueDays:              0,
				RepayStatus:              0,
				RepayTotalShould:         140,
				RepayAmountShould:        0,
				RepayInterestShould:      0,
				RepayFeeShould:           140,
				RepayGraceInterestShould: 0,
				RepayPenaltyShould:       0,
				ForfeitPenalty:           0,
				RepayedDate:              1527782400000,
				RepayedTotal:             140,
				RepayedAmount:            0,
				RepayedInterest:          0,
				RepayedGraceInterest:     0,
				RepayedFee:               140,
				RepayedPenalty:           0,
				RepayedForfeitPenalty:    0,
			},
			// 1期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          1,
				RepayDateShould:          1528387200000,
				Loan:                     0,
				OverdueDays:              94,
				RepayStatus:              int(types.LoanStatusOverdue),
				RepayTotalShould:         14280,
				RepayAmountShould:        5000,
				RepayInterestShould:      350,
				RepayFeeShould:           0,
				RepayGraceInterestShould: 50,
				RepayPenaltyShould:       8880,
				ForfeitPenalty:           0,
				RepayedDate:              1535731200000,
				RepayedTotal:             9000,
				RepayedAmount:            200,
				RepayedInterest:          350,
				RepayedGraceInterest:     50,
				RepayedFee:               0,
				RepayedPenalty:           8400,
				RepayedForfeitPenalty:    0,
			},
		},
			nil},

		// 当前时间还款期外大于90天 部分还款在90天外 未还清
		{types.ProductTrialCalcIn{
			ID:           180508030000776897,
			Loan:         0,
			Amount:       5000,
			Period:       7,
			LoanDate:     "2018-06-01",
			CurrentDate:  "2018-09-10",
			RepayDate:    "2018-09-09",
			RepayedTotal: 9000,
		}, []types.ProductTrialCalcResult{
			// 0期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          0,
				RepayDateShould:          1527782400000,
				Loan:                     4860,
				OverdueDays:              0,
				RepayStatus:              0,
				RepayTotalShould:         140,
				RepayAmountShould:        0,
				RepayInterestShould:      0,
				RepayFeeShould:           140,
				RepayGraceInterestShould: 0,
				RepayPenaltyShould:       0,
				ForfeitPenalty:           0,
				RepayedDate:              1527782400000,
				RepayedTotal:             140,
				RepayedAmount:            0,
				RepayedInterest:          0,
				RepayedGraceInterest:     0,
				RepayedFee:               140,
				RepayedPenalty:           0,
				RepayedForfeitPenalty:    0,
			},
			// 1期
			types.ProductTrialCalcResult{
				NumberOfPeriods:          1,
				RepayDateShould:          1528387200000,
				Loan:                     0,
				OverdueDays:              94,
				RepayStatus:              int(types.LoanStatusOverdue),
				RepayTotalShould:         14300,
				RepayAmountShould:        5000,
				RepayInterestShould:      350,
				RepayFeeShould:           0,
				RepayGraceInterestShould: 50,
				RepayPenaltyShould:       8900,
				ForfeitPenalty:           0,
				RepayedDate:              1536422400000,
				RepayedTotal:             9000,
				RepayedAmount:            0,
				RepayedInterest:          50,
				RepayedGraceInterest:     50,
				RepayedFee:               0,
				RepayedPenalty:           8900,
				RepayedForfeitPenalty:    0,
			},
		},
			nil},
	}

	product := models.Product{
		Id:                 180508030000776897,
		Period:             types.ProductPeriodOfDay,
		DayInterestRate:    100,
		DayFeeRate:         40,
		DayGraceRate:       100,
		DayPenaltyRate:     200,
		ChargeInterestType: types.ProductChargeInterestTypeByStages,
		ChargeFeeType:      types.ProductChargeFeeInterestBefore,
		RepayOrder:         "z;f;g;s;i;p",
		RepayType:          types.ProductRepayTypeOnce,
		CeilWay:            types.ProductCeilWayUp,
		CeilWayUnit:        types.ProductCeilWayUnitOne,
		GracePeriod:        1,
	}
	for _, te := range test {

		// product, _ := models.GetProduct(te.trialIn.ID)
		results, _ := TrialCalcRepayTypeOnce(te.trialIn, product)

		for k, result := range results {
			resultExp := te.trialResults[k]

			gotError := false
			if result.RepayStatus != resultExp.RepayStatus {
				t.Errorf("RepayStatus not match  result.RepayStatus :%d resultExp.RepayStatus %d  ",
					result.RepayStatus, resultExp.RepayStatus)
				gotError = true
			}

			if result.NumberOfPeriods != resultExp.NumberOfPeriods {
				t.Errorf("NumberOfPeriods not match  result.NumberOfPeriods :%d resultExp.NumberOfPeriods %d  ",
					result.NumberOfPeriods, resultExp.NumberOfPeriods)
				gotError = true
			}

			if result.RepayDateShould != resultExp.RepayDateShould {
				t.Errorf("RepayDateShould not match  result.RepayDateShould :%d resultExp.RepayDateShould %d ",
					result.RepayDateShould, resultExp.RepayDateShould)
				gotError = true

			}

			if result.Loan != resultExp.Loan {
				t.Errorf("Loan not match  result.Loan :%d resultExp.Loan %d ",
					result.Loan, resultExp.Loan)
				gotError = true

			}

			if result.OverdueDays != resultExp.OverdueDays {
				t.Errorf("OverdueDays not match  result.OverdueDays :%d resultExp.OverdueDays %d ",
					result.OverdueDays, resultExp.OverdueDays)
				gotError = true

			}

			if result.RepayTotalShould != resultExp.RepayTotalShould {
				t.Errorf("RepayTotalShould not match  result.RepayTotalShould :%d resultExp.RepayTotalShould %d ",
					result.RepayTotalShould, resultExp.RepayTotalShould)
				gotError = true

			}

			if result.RepayAmountShould != resultExp.RepayAmountShould {
				t.Errorf("RepayAmountShould not match  result.RepayAmountShould :%d resultExp.RepayAmountShould %d ",
					result.RepayAmountShould, resultExp.RepayAmountShould)
				gotError = true

			}

			if result.RepayInterestShould != resultExp.RepayInterestShould {
				t.Errorf("RepayInterestShould not match  result.RepayInterestShould :%d resultExp.RepayInterestShould %d ",
					result.RepayInterestShould, resultExp.RepayInterestShould)
				gotError = true

			}

			if result.RepayFeeShould != resultExp.RepayFeeShould {
				t.Errorf("RepayFeeShould not match  result.RepayFeeShould :%d resultExp.RepayFeeShould %d ",
					result.RepayFeeShould, resultExp.RepayFeeShould)
				gotError = true

			}

			if result.RepayGraceInterestShould != resultExp.RepayGraceInterestShould {
				t.Errorf("RepayGraceInterestShould not match  result.RepayGraceInterestShould :%d resultExp.RepayGraceInterestShould %d ",
					result.RepayGraceInterestShould, resultExp.RepayGraceInterestShould)
				gotError = true

			}

			if result.RepayPenaltyShould != resultExp.RepayPenaltyShould {
				t.Errorf("RepayPenaltyShould not match  result.RepayPenaltyShould :%d resultExp.RepayPenaltyShould %d ",
					result.RepayPenaltyShould, resultExp.RepayPenaltyShould)
				gotError = true

			}

			if result.ForfeitPenalty != resultExp.ForfeitPenalty {
				t.Errorf("ForfeitPenalty not match  result.ForfeitPenalty :%d resultExp.ForfeitPenalty %d ",
					result.ForfeitPenalty, resultExp.ForfeitPenalty)
				gotError = true

			}

			if result.RepayedDate != resultExp.RepayedDate {
				t.Errorf("RepayedDate not match  result.RepayedDate :%d resultExp.RepayedDate %d k %d",
					result.RepayedDate, resultExp.RepayedDate, k)
				gotError = true

			}
			if result.RepayedTotal != resultExp.RepayedTotal {
				t.Errorf("RepayedTotal not match  result.RepayedTotal :%d resultExp.RepayedTotal %d k %d",
					result.RepayedTotal, resultExp.RepayedTotal, k)
				gotError = true

			}
			if result.RepayedAmount != resultExp.RepayedAmount {
				t.Errorf("RepayedAmount not match  result.RepayedAmount :%d resultExp.RepayedAmount %d ",
					result.RepayedAmount, resultExp.RepayedAmount)
				gotError = true

			}
			if result.RepayedInterest != resultExp.RepayedInterest {
				t.Errorf("RepayedInterest not match  result.RepayedInterest :%d resultExp.RepayedInterest %d ",
					result.RepayedInterest, resultExp.RepayedInterest)
				gotError = true

			}
			if result.RepayedGraceInterest != resultExp.RepayedGraceInterest {
				t.Errorf("RepayedGraceInterest not match  result.RepayedGraceInterest :%d resultExp.RepayedGraceInterest %d ",
					result.RepayedGraceInterest, resultExp.RepayedGraceInterest)
				gotError = true

			}
			if result.RepayedFee != resultExp.RepayedFee {
				t.Errorf("RepayedFee not match  result.RepayedFee :%d resultExp.RepayedFee %d k %d",
					result.RepayedFee, resultExp.RepayedFee, k)
				gotError = true

			}
			if result.RepayedPenalty != resultExp.RepayedPenalty {
				t.Errorf("RepayedPenalty not match  result.RepayedPenalty :%d resultExp.RepayedPenalty %d ",
					result.RepayedPenalty, resultExp.RepayedPenalty)
				gotError = true

			}

			if result.RepayedForfeitPenalty != resultExp.RepayedForfeitPenalty {
				t.Errorf("RepayedForfeitPenalty not match  result.RepayedForfeitPenalty :%d resultExp.RepayedForfeitPenalty %d ",
					result.RepayedForfeitPenalty, resultExp.RepayedForfeitPenalty)
				gotError = true

			}

			if gotError {
				logs.Debug("=====================")
				logs.Debug("result[%d]  %#v ", k, result)
				logs.Debug("resultExp[%d]  %#v ", k, resultExp)
			}
		}

		// reflect.TypeOf(o) != reflect.TypeOf(d.out) {
		// 	t.Errorf("[SMS delivery] key [%s]对应的预设sender类型[%v] 与实际返回类型[%v]不符 ",
		// 		d.in, reflect.TypeOf(d.out), reflect.TypeOf(o))
		// }
	}
}

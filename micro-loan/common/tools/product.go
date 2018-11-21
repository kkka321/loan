package tools

import (
	"math"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
)

func CeilWay(value float64, feeBase int64, way types.ProductCeilWayEunm, unit types.ProductCeilWayUnitEunm) (result int64) {
	switch way {
	case types.ProductCeilWayUp:
		{
			result = int64(math.Ceil(float64(value)/float64(feeBase)/float64(unit))) * int64(unit)
		}
	case types.ProductCeilWayNo:
		{
			result = int64(value)
		}
	default:
		{
			logs.Warn("undefine ceilway:", way)
		}
	}
	return
}

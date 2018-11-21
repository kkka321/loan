package performance

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/types"
	"strconv"
	"strings"

	"github.com/astaxie/beego/orm"
)

// ItemStatsListBackend 返回
func ItemStatsListBackend(condCntr map[string]interface{}, page int, pagesize int) (list []models.TicketItemMonthlyStats, total int64, err error) {
	obj := models.TicketItemMonthlyStats{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = types.DefaultPagesize
	}
	offset := (page - 1) * pagesize

	// 初始化查询条件
	where := itemStatsWhereBackend(condCntr)

	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` %s", obj.TableName(), where)
	sqlList := fmt.Sprintf("SELECT * FROM `%s` %s ORDER BY `date` DESC,ticket_item_id ASC LIMIT %d,%d", obj.TableName(), where, offset, pagesize)

	// 查询符合条件的所有条数
	r := o.Raw(sqlCount)
	r.QueryRow(&total)

	// 查询指定页
	r = o.Raw(sqlList)
	r.QueryRows(&list)

	//

	return
}

func itemStatsWhereBackend(condCntr map[string]interface{}) string {
	// 初始化查询条件
	cond := []string{}

	if v, ok := condCntr["ticket_item_id"]; ok {
		if itemID, ok := v.(types.TicketItemEnum); ok {
			cond = append(cond, fmt.Sprintf("ticket_item_id=%d", itemID))
		} else if itemIDMap, ok := v.(map[types.TicketItemEnum]string); ok {
			// for manager
			var itemIDs []string
			for k := range itemIDMap {
				itemIDs = append(itemIDs, strconv.Itoa(int(k)))
			}
			cond = append(cond, fmt.Sprintf("ticket_item_id in(%s)", strings.Join(itemIDs, ",")))
		}
	}

	if len(cond) > 0 {
		return "WHERE " + strings.Join(cond, " AND ")
	}
	return ""
}

package performance

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/pkg/rbac"
	"micro-loan/common/types"
	"strconv"
	"strings"

	"github.com/astaxie/beego/orm"
)

// WorkerStatsListBackend 返回
func WorkerStatsListBackend(condCntr map[string]interface{}, roleID int64, selfUID int64, page int, pagesize int) (list []models.TicketWorkerDailyStats, total int64, totalStats WorkerStatsTotal, err error) {
	obj := models.TicketWorkerDailyStats{}
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
	where := workerStatsWhereBackend(condCntr, roleID, selfUID)

	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` %s", obj.TableName(), where)
	sqlList := fmt.Sprintf("SELECT * FROM `%s` %s ORDER BY `date` DESC,ticket_item_id ASC, ranking ASC LIMIT %d,%d", obj.TableName(), where, offset, pagesize)

	// 查询符合条件的所有条数
	r := o.Raw(sqlCount)
	r.QueryRow(&total)

	// 查询指定页
	r = o.Raw(sqlList)
	r.QueryRows(&list)

	//
	totalStats = workerStatsTotalBackend(where)

	return
}

// WorkerStatsListExportBackend 返回
func WorkerStatsListExportBackend(condCntr map[string]interface{}, roleID int64, selfUID int64) (list []models.TicketWorkerDailyStats, total int64, totalStats WorkerStatsTotal, err error) {
	obj := models.TicketWorkerDailyStats{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	// 初始化查询条件
	where := workerStatsWhereBackend(condCntr, roleID, selfUID)

	// sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` %s", obj.TableName(), where)
	sqlList := fmt.Sprintf("SELECT * FROM `%s` %s ORDER BY `date` DESC,ticket_item_id ASC, ranking ASC", obj.TableName(), where)

	// 查询符合条件的所有条数
	// r := o.Raw(sqlCount)
	// r.QueryRow(&total)

	// 查询指定页
	r := o.Raw(sqlList)
	r.QueryRows(&list)

	//
	totalStats = workerStatsTotalBackend(where)

	return
}

// WorkerHourStatsListBackend 返回
func WorkerHourStatsListBackend(condCntr map[string]interface{}, roleID int64, rolePid int64, page int, pagesize int) (list []models.TicketWorkerHourlyStats, total int64, totalStats WorkerStatsTotal, err error) {
	obj := models.TicketWorkerHourlyStats{}
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
	where := workerHourStatsWhereBackend(condCntr, roleID, rolePid)

	sqlCount := fmt.Sprintf("SELECT COUNT(`id`) FROM `%s` %s", obj.TableName(), where)
	sqlList := fmt.Sprintf("SELECT * FROM `%s` %s ORDER BY `hour` DESC,ticket_item_id ASC, ranking ASC LIMIT %d,%d", obj.TableName(), where, offset, pagesize)

	// 查询符合条件的所有条数
	r := o.Raw(sqlCount)
	r.QueryRow(&total)

	// 查询指定页
	r = o.Raw(sqlList)
	r.QueryRows(&list)

	//
	totalStats = workerHourStatsTotalBackend(where)

	return
}

// WorkerStatsTotal 描述总计
type WorkerStatsTotal struct {
	TotalAssign int64
	TotalHandle int64
}

// workerStatsTotalBackend 返回
func workerStatsTotalBackend(where string) (result WorkerStatsTotal) {
	obj := models.TicketWorkerDailyStats{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	// 初始化查询条件
	sqlList := fmt.Sprintf("SELECT sum(assign_num) as total_assign,sum(handle_num) as total_handle  FROM `%s` %s", obj.TableName(), where)

	// 查询指定页
	r := o.Raw(sqlList)
	r.QueryRow(&result)

	return
}

// workerHourStatsTotalBackend 返回
func workerHourStatsTotalBackend(where string) (result WorkerStatsTotal) {
	obj := models.TicketWorkerHourlyStats{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	// 初始化查询条件
	sqlList := fmt.Sprintf("SELECT sum(assign_num) as total_assign,sum(handle_num) as total_handle  FROM `%s` %s", obj.TableName(), where)

	// 查询指定页
	r := o.Raw(sqlList)
	r.QueryRow(&result)

	return
}

func workerStatsWhereBackend(condCntr map[string]interface{}, roleID, selfUID int64) string {
	// 初始化查询条件
	cond := []string{}
	if v, ok := condCntr["id"]; ok {
		cond = append(cond, fmt.Sprintf("id=%d", v))
	}

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
	if v, ok := condCntr["admin_uid"]; ok {
		// it means already
		cond = append(cond, fmt.Sprintf("admin_uid=%d", v))
	} else {
		// for manager page
		if rbac.GetRoleLevel(roleID) != types.RoleSuper {
			//if rolePid != types.RoleSuperPid {
			var uids []string
			models.GetUserIDsByRolePidFromDB(roleID, &uids)
			uids = append(uids, strconv.FormatInt(selfUID, 10))
			cond = append(cond, fmt.Sprintf("admin_uid in(%s)", strings.Join(uids, ",")))
		}
	}

	if v, ok := condCntr["date_start"]; ok {
		cond = append(cond, fmt.Sprintf("date>=%d", v))
	}
	if v, ok := condCntr["date_end"]; ok {
		cond = append(cond, fmt.Sprintf("date<=%d", v))
	}

	if len(cond) > 0 {
		return "WHERE " + strings.Join(cond, " AND ")
	}
	return ""
}

func workerHourStatsWhereBackend(condCntr map[string]interface{}, roleID, rolePid int64) string {
	// 初始化查询条件
	cond := []string{}
	if v, ok := condCntr["id"]; ok {
		cond = append(cond, fmt.Sprintf("id=%d", v))
	}

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
	if v, ok := condCntr["admin_uid"]; ok {
		// it means already
		cond = append(cond, fmt.Sprintf("admin_uid=%d", v))
	} else {
		// for manager page
		if rolePid != types.RoleSuperPid {
			var uids []string
			subUserNums, _ := models.GetUserIDsByRolePidFromDB(roleID, &uids)
			if subUserNums == 0 {
				uids = append(uids, "")
			}
			cond = append(cond, fmt.Sprintf("admin_uid in(%s)", strings.Join(uids, ",")))
		}
	}

	if v, ok := condCntr["currentHour"]; ok {
		cond = append(cond, fmt.Sprintf("hour=%d", v))
	}

	if len(cond) > 0 {
		return "WHERE " + strings.Join(cond, " AND ")
	}
	return ""
}

package rbac

import (
	"encoding/json"
	"errors"
	"fmt"
	"micro-loan/common/lib/redis/cache"
	"micro-loan/common/models"
	"micro-loan/common/tools"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

const (
	// RootParentMenuID 菜单根节点 PID 0
	RootParentMenuID = 0
	// MaxMenuDepth 菜单树最大层级 3
	MaxMenuDepth = 3
)

// MenuTreeNode describe a menu node , 包含子节点slice, 用于菜单管理
type MenuTreeNode struct {
	models.Menu
	SubMenus []MenuTreeNode
	IsActive bool
}

// MenuList 返回符合条件的所有菜单
// Menu总数据量不大, 所以不分页
func MenuList(condCntr map[string]interface{}) (list []models.Menu, err error) {
	obj := models.Menu{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	cond := "1=1"
	if f, ok := condCntr["pid"]; ok {
		cond += fmt.Sprintf(" AND `pid` = '%s'", f)
	}
	if f, ok := condCntr["pid"]; ok {
		cond += fmt.Sprintf(" AND `pid` = '%s'", f)
	}

	// 排序不可更改, 后面树的生成依赖排序
	sqlList := fmt.Sprintf("SELECT * FROM `%s` WHERE %s ORDER BY `pid` DESC,`sort` ASC", models.MENU_TABLENAME, cond)

	// 查询指定页
	r := o.Raw(sqlList)
	r.QueryRows(&list)

	return
}

// MenuTree 返回一个完整的菜单树
func MenuTree() (treeList []MenuTreeNode, err error) {
	redisCache := cache.RedisCacheClient.Get()
	defer redisCache.Close()
	cacheKey := beego.AppConfig.String("rbac_menu")
	valBytes, err := redisCache.Do("GET", cacheKey)
	if err == nil && valBytes != nil {
		json.Unmarshal(valBytes.([]byte), &treeList)
		return
	}

	// 从数据库获取menu数据, 并
	{
		allMenuList, _ := MenuList(nil)
		mapSameLayerMenuList := map[int64][]MenuTreeNode{}
		for _, v := range allMenuList {
			var menuNode MenuTreeNode
			if subList, ok := mapSameLayerMenuList[v.Id]; ok {
				// subList 存在, 则将 subList 合入父亲节点
				menuNode = MenuTreeNode{v, subList, false}
				delete(mapSameLayerMenuList, v.Id)
			} else {
				menuNode = MenuTreeNode{v, nil, false}
			}
			if set, ok := mapSameLayerMenuList[v.Pid]; ok {
				mapSameLayerMenuList[v.Pid] = append(set, menuNode)
			} else {
				mapSameLayerMenuList[v.Pid] = []MenuTreeNode{menuNode}
			}
		}
		if len(mapSameLayerMenuList) > 1 {
			err = fmt.Errorf("[MenuList] should be one tree, root node num:%d", len(mapSameLayerMenuList))
			logs.Error(err)
			return
		}

		if v, ok := mapSameLayerMenuList[RootParentMenuID]; ok {
			treeList = v
		} else {
			err = errors.New("No root node")
		}
	}

	// 执行到此处说明已经生成 menu tree
	jsonBytes, _ := json.Marshal(treeList)
	redisCache.Do("SET", cacheKey, jsonBytes)

	return
}

func clearMenuTreeCache() {
	redisCache := cache.RedisCacheClient.Get()
	defer redisCache.Close()

	cacheKey := beego.AppConfig.String("rbac_menu")
	_, err := redisCache.Do("DEL", cacheKey)
	if err != nil {
		logs.Error(err)
	}
}

// AuthMenuTree 返回已授权的 menu tree
// active menu 已经被标记
func AuthMenuTree(activePid int64, pidMap map[int64]bool) (treeList []MenuTreeNode, err error) {
	allMenuTree, _ := MenuTree()
	for _, v := range allMenuTree {
		// 含有子节点
		if len(v.SubMenus) > 0 {
			// 深度拷贝 v 至 topNode
			var topNode MenuTreeNode
			topNode.Class = v.Class
			topNode.Ctime = v.Ctime
			topNode.Id = v.Id
			topNode.Name = v.Name
			topNode.Path = v.Path
			topNode.Pid = v.Pid
			topNode.PrivilegeId = v.PrivilegeId
			topNode.Sort = v.Sort
			topNode.Status = v.Status
			topNode.Utime = v.Utime

			for _, secV := range v.SubMenus {
				// 三级节点
				if len(secV.SubMenus) > 0 {
					var secNode MenuTreeNode
					secNode.Class = secV.Class
					secNode.Ctime = secV.Ctime
					secNode.Id = secV.Id
					secNode.Name = secV.Name
					secNode.Path = secV.Path
					secNode.Pid = secV.Pid
					secNode.PrivilegeId = secV.PrivilegeId
					secNode.Sort = secV.Sort
					secNode.Status = secV.Status
					secNode.Utime = secV.Utime
					for _, thirdV := range secV.SubMenus {
						if _, ok := pidMap[thirdV.PrivilegeId]; ok {
							// active menu
							if activePid == thirdV.PrivilegeId {
								thirdV.IsActive = true
								secNode.IsActive = true
								topNode.IsActive = true
							}
							secNode.SubMenus = append(secNode.SubMenus, thirdV)
						}
					}
					if len(secNode.SubMenus) > 0 {
						topNode.SubMenus = append(topNode.SubMenus, secNode)
					}
				} else {
					if _, ok := pidMap[secV.PrivilegeId]; ok {
						// active menu
						if activePid == secV.PrivilegeId {
							secV.IsActive = true
							topNode.IsActive = true
						}
						topNode.SubMenus = append(topNode.SubMenus, secV)
					}
				}
			}

			if len(topNode.SubMenus) > 0 {
				treeList = append(treeList, topNode)
			}
		} else {
			if _, ok := pidMap[v.PrivilegeId]; ok {
				// active menu
				if activePid == v.PrivilegeId {
					v.IsActive = true
				}
				treeList = append(treeList, v)
			}
		}
	}
	return
}

// SuperMenuTree 返回已授权的 menu tree
// active menu 已经被标记
func SuperMenuTree(activePid int64) (allMenuTree []MenuTreeNode, err error) {
	allMenuTree, _ = MenuTree()
	for i, v := range allMenuTree {
		// 含有子节点
		if len(v.SubMenus) > 0 {
			// 深度拷贝 v 至 topNode

			for secI, secV := range v.SubMenus {
				// 三级节点
				if len(secV.SubMenus) > 0 {

					for thirdI, thirdV := range secV.SubMenus {

						if activePid == thirdV.PrivilegeId {
							allMenuTree[i].SubMenus[secI].SubMenus[thirdI].IsActive = true
							allMenuTree[i].SubMenus[secI].IsActive = true
							allMenuTree[i].IsActive = true
						}
					}

				} else {
					// active menu
					if activePid == secV.PrivilegeId {
						allMenuTree[i].SubMenus[secI].IsActive = true
						allMenuTree[i].IsActive = true

					}
				}
			}

		} else {
			if activePid == v.PrivilegeId {
				allMenuTree[i].IsActive = true
			}
		}
	}
	return
}

// ActiveMenuTreeForUpdate 返回已授权的 menu tree
// active menu 已经被标记
func ActiveMenuTreeForUpdate(id int64) (allMenuTree []MenuTreeNode, err error) {
	allMenuTree, _ = MenuTree()
	for i, v := range allMenuTree {
		// 含有子节点
		if id == v.Id {
			allMenuTree[i].IsActive = true
			return
		}
		if len(v.SubMenus) > 0 {
			// 深度拷贝 v 至 topNode

			for secI, secV := range v.SubMenus {
				// 三级节点
				if id == secV.Id {
					allMenuTree[i].SubMenus[secI].IsActive = true
					allMenuTree[i].IsActive = true
					return
				}
				if len(secV.SubMenus) > 0 {
					for thirdI, thirdV := range secV.SubMenus {
						if id == thirdV.Id {
							allMenuTree[i].SubMenus[secI].SubMenus[thirdI].IsActive = true
							allMenuTree[i].SubMenus[secI].IsActive = true
							allMenuTree[i].IsActive = true
							return
						}
					}

				}
			}

		}
	}
	return
}

func getMaxBrotherSort(pid int64) (sort int) {
	obj := models.Menu{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	// 排序不可更改, 后面树的生成依赖排序
	sqlList := fmt.Sprintf("SELECT sort FROM `%s` WHERE pid=%d ORDER BY `sort` DESC limit 1", models.MENU_TABLENAME, pid)

	// 查询指定页
	r := o.Raw(sqlList)
	r.QueryRow(&sort)

	return
}

// AddOneMenu 新增 Menu
// 含属性校验, 外部可直接调用
// 内部自动初始化, Ctime 与 Utime
func AddOneMenu(m *models.Menu) (id int64, err error) {
	m.Sort = getMaxBrotherSort(m.Pid) + 1

	o := orm.NewOrm()
	o.Using(m.Using())
	//
	m.Ctime = tools.GetUnixMillis()
	m.Utime = m.Ctime
	id, err = o.Insert(m)
	if id > 0 {
		clearMenuTreeCache()
	}
	return
}

// UpdateOneMenu 更新指定Menu的属性
// 不含属性校验
// 内部自动更新 Utime
func UpdateOneMenu(m *models.Menu, cols ...string) (affectedRows int64, err error) {

	if m.Id <= 0 {
		err = errors.New("Update ID must exist and >0")
		return
	}

	m.Utime = tools.GetUnixMillis()
	o := orm.NewOrm()
	o.Using(m.Using())
	affectedRows, err = o.Update(m, cols...)
	if affectedRows > 0 {
		clearMenuTreeCache()
	}
	return
}

// DeleteOneMenu 删除指定 menu, 非父级menu
func DeleteOneMenu(m *models.Menu) (affectedRows int64, err error) {

	if m.Id <= 0 {
		err = errors.New("Update ID must exist and >0")
		return
	}

	o := orm.NewOrm()
	o.Using(m.Using())

	var subNum int
	subMenuSQLCount := fmt.Sprintf("SELECT COUNT(*) FROM `%s` WHERE pid=%d ", models.MENU_TABLENAME, m.Id)
	r := o.Raw(subMenuSQLCount)
	r.QueryRow(&subNum)
	if subNum > 0 {
		err = fmt.Errorf("当前菜单有 %d 个子菜单, 请先删除子菜单", subNum)
		return
	}

	affectedRows, err = o.Delete(m)
	if affectedRows > 0 {
		clearMenuTreeCache()
	}
	return
}

// UpdateMenuSort 更新指定Menu的属性
// 不含属性校验
// 内部自动更新 Utime
// UpdateMenuSort 更新指定Menu的属性
// 不含属性校验
// 内部自动更新 Utime
func UpdateMenuSort(m *models.Menu, Up bool) (affectedRows int64, err error) {
	brotherM := &models.Menu{}

	o := orm.NewOrm()
	o.Using(brotherM.Using())
	// 读取当前menu
	o.QueryTable(m.TableName()).Filter("id", m.Id).One(m)

	// 构造兄弟节点查询条件
	cond := fmt.Sprintf("pid=%d", m.Pid)
	var orderBy string
	if Up {
		// Get big brother menu
		cond += fmt.Sprintf(" AND sort<%d", m.Sort)
		orderBy = "`sort` DESC"
	} else {
		// Get little brother menu
		cond += fmt.Sprintf(" AND sort>%d", m.Sort)
		orderBy = "`sort` ASC"
	}

	sql := fmt.Sprintf("SELECT * FROM `%s` WHERE %s ORDER BY %s limit 1", models.MENU_TABLENAME, cond, orderBy)
	// 执行兄弟节点查询
	r := o.Raw(sql)
	r.QueryRow(&brotherM)
	if brotherM.Id <= 0 {
		// 无上节点或下节点, 无法移动
		err = errors.New("菜单已经位于顶部或者尾部")
		return
	}

	// 开始置换 sort 事务
	o.Begin()
	m.Sort, brotherM.Sort = brotherM.Sort, m.Sort
	affectedRows1, _ := o.Update(m, "Sort")
	affectedRows2, _ := o.Update(brotherM, "Sort")

	err = o.Commit()
	if err != nil || affectedRows1 != 1 || affectedRows2 != 1 {
		// 失败回滚
		o.Rollback()
		err = errors.New("操作失败, 请重试")
		logs.Error(m, brotherM)
		return
	}
	// 成功返回,影响行数
	affectedRows = affectedRows1 + affectedRows2
	clearMenuTreeCache()

	return
}

// GetOneMenu 获取指定ID的角色信息
func GetOneMenu(id int64) (data models.Menu, err error) {
	obj := models.Menu{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())
	qs := o.QueryTable(obj.TableName())

	err = qs.Filter("id", id).One(&data)

	return
}

// todo recursive match for menu
// func privilegeMatch(pList *[]MenuTreeNode, treeList *[]MenuTreeNode, pidMap *map[int64]bool) {
// 	for i, v := range MenuTreeNode {
//
// 		if len(v.SubMenus) > 0 {
// 			var node MenuTreeNode
// 			node.Class = v.Class
// 			node.Ctime = v.Ctime
// 			node.Id = v.Id
// 			node.Name = v.Name
// 			node.Path = v.Path
// 			node.Pid = v.Pid
// 			node.PrivilegeId = v.PrivilegeId
// 			node.Sort = v.Sort
// 			node.Status = v.Status
// 			node.Utime = v.Utime
//
// 			//privlegeMatch(&pList, &node.SubMenus, &pidMap);
// 			for _, secV := range v.SubMenus {
// 				if _, ok := pidMap[secV.PrivilegeId]; ok {
// 					node.SubMenus = append(node.SubMenus, secV)
// 				}
// 			}
//
// 			if len(node.SubMenus) > 0 {
// 				treeList = append(treeList, topNode)
// 			}
// 		} else {
// 			if _, ok := pidMap[v.PrivilegeId]; ok {
// 				treeList = append(treeList, v)
// 			}
// 		}
// 	}
// }

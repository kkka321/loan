package rbac

// DynamicResource 动态资源
type DynamicResource struct {
	ID        int64
	Name      int
	RelatedID int64
	OwnerType int
	OwnerID   int64
	// 1 拥有, 回收
	Status int
}

// 资源和工单是否要交互

// select * from resource WHERE type="customer" AND
// (OwnerID in uids)

// 超管拥有所有资源
// 管理者拥有本身uid, 和 管理uids 下所有资源

// 用户拥有资源=用户自身动态资源+角色动态资源权限+子用户资源权限集合
// 简单来说: 用户拥有资源分为三部分, 一是管辖资源, 二是继承资源, 三是特有用户分配资源
// 上级角色默认资源集 > 下级角色默认资源集 (部分拷贝关系)
// SELECT * FROM resource WHERE type="customer" AND
// ((OwnerID in(self_uid, sub_uids[])  AND OwnerType = "user")
// OR (owner_id = roleID AND  OwnerType = "role") ) and status = 1

//

func (res *DynamicResource) ReadAll() {

}

func (res *DynamicResource) Grant() {

}

func (res *DynamicResource) Revoke() {

}

func (res *DynamicResource) GetResByUID() {

}

func (res *DynamicResource) GetResByRoleID() {

}

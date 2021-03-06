package roleapp

import (
	"fmt"
	. "github.com/leyle/ginbase/consolelog"
	"github.com/leyle/ginbase/dbandmq"
	"github.com/leyle/ginbase/util"
	"github.com/leyle/userandrole/ophistory"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// role -> permissions -> items
// 简单处理，不允许继承

// 程序启动时，初始化出来的
const DefaultRoleName = "注册用户默认角色"

var DefaultRoleId = ""

const (
	AdminRoleName       = "admin"
	AdminPermissionName = "admin"
	AdminItemName       = "admin:"
)

// 定义不可修改的 item / permissid / role Id
const (
	IdTypeItem       = "item"
	IdTypePermission = "permission"
	IdTypeRole       = "role"
)

var (
	CanNotModifyItemIds       []string
	CanNotModifyPermissionIds []string
	CanNotMidifyRoleIds       []string
)

var AdminItemNames = []string{
	AdminItemName + "GET",
	AdminItemName + "POST",
	AdminItemName + "PUT",
	AdminItemName + "DELETE",
	AdminItemName + "PATCH",
	AdminItemName + "OPTION",
	AdminItemName + "HEAD",
}

// 数据来源，一种是系统内置，一种是用户定义，用户定义可以自定义标签，也可以使用默认值
const (
	DataFromSystem = "SYSTEM" // 系统内置数据
	DataFromUser   = "USER"   // 用户传递的
)

// item
const CollectionNameItem = "item"

var IKItem = &dbandmq.IndexKey{
	Collection: CollectionNameItem,
	SingleKey:  []string{"name", "method", "path", "deleted"},
	UniqueKey:  []string{"name"},
}

type Item struct {
	Id   string `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`
	// api
	Method   string `json:"method" bson:"method"`
	Path     string `json:"path" bson:"path"`
	Resource string `json:"resource" bson:"resource"` // 可以为空

	// html
	Menu   string `json:"menu" bson:"menu"`
	Button string `json:"button" bson:"button"`

	Deleted bool `json:"deleted" bson:"deleted"`

	DataFrom string `json:"-" bson:"dataFrom"`

	History []*ophistory.OperationHistory `json:"history" bson:"history"`

	CreateT *util.CurTime `json:"-" bson:"createT"`
	UpdateT *util.CurTime `json:"-" bson:"updateT"`
}

// permission
const CollectionNamePermission = "permission"

var IKPermission = &dbandmq.IndexKey{
	Collection: CollectionNamePermission,
	SingleKey:  []string{"name", "itemIds", "deleted"},
	UniqueKey:  []string{"name"},
}

type Permission struct {
	Id   string `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`

	ItemIds []string `json:"-" bson:"itemIds"`
	Items   []*Item  `json:"items" bson:"-"`

	// html
	Menu   string `json:"menu" bson:"menu"`
	Button string `json:"button" bson:"button"`

	Deleted  bool   `json:"deleted" bson:"deleted"`
	DataFrom string `json:"-" bson:"dataFrom"`

	History []*ophistory.OperationHistory `json:"history" bson:"history"`

	CreateT *util.CurTime `json:"-" bson:"createT"`
	UpdateT *util.CurTime `json:"-" bson:"updateT"`
}

// role
const CollectionNameRole = "role"

var IKRole = &dbandmq.IndexKey{
	Collection: CollectionNameRole,
	SingleKey:  []string{"name", "permissionIds", "deleted"},
	UniqueKey:  []string{"name"},
}

type Role struct {
	Id   string `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`

	PermissionIds []string      `json:"-" bson:"permissionIds"`
	Permissions   []*Permission `json:"permissions" bson:"-"`

	// html
	Menu   string `json:"menu" bson:"menu"`
	Button string `json:"button" bson:"button"`

	// 包含的下属 role 列表，当前 role 所属用户可以给自己的下属用户赋予的权限
	ChildrenRoles []*ChildRole `json:"childrenRole" bson:"childrenRole"`

	Deleted  bool   `json:"deleted" bson:"deleted"`
	DataFrom string `json:"-" bson:"dataFrom"`

	History []*ophistory.OperationHistory `json:"history" bson:"history"`

	CreateT *util.CurTime `json:"-" bson:"createT"`
	UpdateT *util.CurTime `json:"-" bson:"updateT"`
}

// 记录值，归属于某个上层 role
type ChildRole struct {
	Id   string `json:"id" bson:"id"`     // role Id
	Name string `json:"name" bson:"name"` // role name，展示查看用
}

// 根据 id 读取 item
func GetItemById(db *dbandmq.Ds, id string) (*Item, error) {
	var item *Item
	err := db.C(CollectionNameItem).FindId(id).One(&item)
	if err != nil && err != mgo.ErrNotFound {
		Logger.Errorf("", "根据id[%s]读取 item 信息失败, %s", id, err.Error())
		return nil, err
	}
	return item, nil
}

// 根据 name 读取 item
func GetItemByName(db *dbandmq.Ds, name string) (*Item, error) {
	f := bson.M{
		"name": name,
	}

	var item *Item
	err := db.C(CollectionNameItem).Find(f).One(&item)
	if err != nil && err != mgo.ErrNotFound {
		Logger.Errorf("", "根据name[%s]读取 role item 失败, %s", err.Error())
		return nil, err
	}

	return item, nil
}

// 存储 item
func SaveItem(db *dbandmq.Ds, item *Item) error {
	return db.C(CollectionNameItem).Insert(item)
}

// 更新指定 id 的 item
func UpdateItem(db *dbandmq.Ds, item *Item) error {
	err := db.C(CollectionNameItem).UpdateId(item.Id, item)
	return err
}

// 删除指定 id 的 item
// 不需要单独的去删除包含了自己的 permission 中的数据
// permission 中会标记这个数据，并且不做显示
func DeleteItemById(db *dbandmq.Ds, userId, userName, id string) error {
	opAction := fmt.Sprintf("删除 item, itemId[%s]", id)
	opHis := ophistory.NewOpHistory(userId, userName, opAction)

	update := bson.M{
		"$set": bson.M{
			"deleted": true,
			"updateT": util.GetCurTime(),
		},
		"$push": bson.M{
			"history": opHis,
		},
	}

	err := db.C(CollectionNameItem).UpdateId(id, update)
	if err != nil {
		Logger.Errorf("", "删除item[%s]失败,%s", id, err.Error())
		return err
	}
	return nil
}

// 根据 name 读取 permission
func GetPermissionByName(db *dbandmq.Ds, name string, more bool) (*Permission, error) {
	f := bson.M{
		"name": name,
	}

	var p *Permission
	err := db.C(CollectionNamePermission).Find(f).One(&p)
	if err != nil && err != mgo.ErrNotFound {
		Logger.Errorf("", "根据permission name[%s]读取permission信息失败, %s", name, err.Error())
		return nil, err
	}

	if p == nil {
		return nil, nil
	}

	if more {
		items, err := GetItemsByItemIds(db, p.ItemIds)
		if err == nil {
			p.Items = items
		}
	}

	return p, nil
}

func GetPermissionById(db *dbandmq.Ds, id string, more bool) (*Permission, error) {
	var p *Permission
	err := db.C(CollectionNamePermission).FindId(id).One(&p)
	if err != nil && err != mgo.ErrNotFound {
		Logger.Errorf("", "根据 permission id[%s]读取permission信息失败, %s", id, err.Error())
		return nil, err
	}

	if p == nil {
		return nil, nil
	}

	if more {
		items, err := GetItemsByItemIds(db, p.ItemIds)
		if err == nil {
			p.Items = items
		}
	}

	return p, nil
}

// 存储 permission
func SavePermission(db *dbandmq.Ds, p *Permission) error {
	return db.C(CollectionNamePermission).Insert(p)
}

func UpdatePermission(db *dbandmq.Ds, p *Permission) error {
	return db.C(CollectionNamePermission).UpdateId(p.Id, p)
}

// 根据 name 读取 role
func GetRoleByName(db *dbandmq.Ds, name string, more bool) (*Role, error) {
	f := bson.M{
		"name": name,
	}

	var role *Role
	err := db.C(CollectionNameRole).Find(f).One(&role)
	if err != nil && err != mgo.ErrNotFound {
		Logger.Errorf("", "根据role name[%s]读取role信息失败, %s", name, err.Error())
		return nil, err
	}

	if role == nil {
		return nil, nil
	}

	if more {
		ps, err := GetPermissionsByPermissionIds(db, role.PermissionIds)
		if err == nil {
			role.Permissions = ps
		}
	}

	return role, nil
}

func GetRoleById(db *dbandmq.Ds, id string, more bool) (*Role, error) {
	var role *Role
	err := db.C(CollectionNameRole).FindId(id).One(&role)
	if err != nil && err != mgo.ErrNotFound {
		Logger.Errorf("", "根据role id[%s]读取role信息失败, %s", id, err.Error())
		return nil, err
	}

	if role == nil {
		return nil, nil
	}

	if more {
		ps, err := GetPermissionsByPermissionIds(db, role.PermissionIds)
		if err == nil {
			role.Permissions = ps
		}
	}

	return role, nil
}

func SaveRole(db *dbandmq.Ds, role *Role) error {
	return db.C(CollectionNameRole).Insert(role)
}

func UpdateRole(db *dbandmq.Ds, role *Role) error {
	return db.C(CollectionNameRole).UpdateId(role.Id, role)
}

func GetFilterItems(db *dbandmq.Ds, filter *bson.M) ([]*Item, error) {
	var items []*Item
	err := db.C(CollectionNameItem).Find(filter).All(&items)
	if err != nil {
		Logger.Errorf("", "查询筛选的 items 失败, %s", err.Error())
		return nil, err
	}

	return items, nil
}

func GetFilterPermissions(db *dbandmq.Ds, filter *bson.M) ([]*Permission, error) {
	var ps []*Permission
	err := db.C(CollectionNamePermission).Find(filter).All(&ps)
	if err != nil {
		Logger.Errorf("", "查询筛选的 permissions 失败, %s", err.Error())
		return nil, err
	}
	return ps, nil
}

func GetFilterRoles(db *dbandmq.Ds, filter *bson.M) ([]*Role, error) {
	var roles []*Role
	err := db.C(CollectionNameRole).Find(filter).All(&roles)
	if err != nil {
		Logger.Errorf("", "查询筛选 roles 失败, %s", err.Error())
		return nil, err
	}
	return roles, nil
}
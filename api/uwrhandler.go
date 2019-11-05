package api

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase/dbandmq"
	"github.com/leyle/ginbase/middleware"
	"github.com/leyle/ginbase/returnfun"
	"github.com/leyle/ginbase/util"
	"github.com/leyle/userandrole/ophistory"
	"github.com/leyle/userandrole/userandrole"
	"github.com/leyle/userandrole/userapp"
	"gopkg.in/mgo.v2/bson"
)

// uwr means user with role

// 给用户添加 roles
type AddRolesToUserForm struct {
	UserId string `json:"userId" binding:"required"`
	RoleIds []string `json:"roleIds" binding:"required"`
}
func AddRolesToUserHandler(c *gin.Context, ds *dbandmq.Ds) {
	var form AddRolesToUserForm
	err := c.BindJSON(&form)
	middleware.StopExec(err)

	db := ds.CopyDs()
	defer db.Close()

	// 检查所有的 roleids 的有效性
	curUser, curRoles := GetCurUserAndRole(c)
	if curUser == nil {
		middleware.StopExec(errors.New("获取当前用户信息失败"))
		return
	}

	if !shareRoleIsValid(curUser, curRoles, form.RoleIds) {
		returnfun.Return403Json(c, "当前用户无权给用户赋予某些权限")
		return
	}

	// 不用锁定数据，低频操作
	uwr, err := addRoleToUser(db, curUser, form.UserId, form.RoleIds)
	middleware.StopExec(err)

	returnfun.ReturnOKJson(c, uwr)
	return
}

func addRoleToUser(db *dbandmq.Ds, curUser *userapp.User, userId string, roleIds []string) (*userandrole.UserWithRole, error) {
	// 检查 uwr 是否存在，不存在新建，存在就是更新
	uwr, err := userandrole.GetUserWithRoleByUserId(db, userId)
	if err != nil {
		return nil, err
	}
	update := true
	if uwr == nil {
		update = false
		uwr = &userandrole.UserWithRole{
			Id:       util.GenerateDataId(),
			UserId:   userId,
			RoleIds:  roleIds,
			CreateT:  util.GetCurTime(),
		}
		uwr.UpdateT = uwr.CreateT
	} else {
		uwr.RoleIds = append(uwr.RoleIds, roleIds...)
	}
	uwr.RoleIds = util.UniqueStringArray(uwr.RoleIds)

	opAction := fmt.Sprintf("给用户[%s]添加roleIds %s", userId, roleIds)
	opHis := ophistory.NewOpHistory(curUser.Id, curUser.Name, opAction)
	uwr.History = append(uwr.History, opHis)

	err = userandrole.SaveUserWithRole(db, uwr, update)
	if err != nil {
		return nil, err
	}

	// 同步追加操作记录到对应的用户上
	_ = userapp.AppendOpHistoryToUser(db, userId, opHis)

	return uwr, nil
}

// 取消 用户的某些 roles
type RemoveRolesFromUserForm struct {
	UserId string `json:"userId" binding:"required"`
	RoleIds []string `json:"roleIds" binding:"required"`
}
func RemoveRolesFromUserHandler(c *gin.Context, ds *dbandmq.Ds) {
	var form RemoveRolesFromUserForm
	err := c.BindJSON(&form)
	middleware.StopExec(err)

	db := ds.CopyDs()
	defer db.Close()

	// 检查 roleids 有效性 todo
	curUser, curRoles := GetCurUserAndRole(c)
	if !shareRoleIsValid(curUser, curRoles, form.RoleIds) {
		returnfun.Return403Json(c, "当前用户无权限删除用户某些权限")
		return
	}

	uwr, err := userandrole.GetUserWithRoleByUserId(db, form.UserId)
	middleware.StopExec(err)
	if uwr == nil {
		returnfun.ReturnErrJson(c, "用户未有授权记录")
		return
	}

	var remainIds []string
	for _, dbr := range uwr.RoleIds {
		remain := true
		for _, rid := range form.RoleIds {
			if dbr == rid {
				remain = false
				break
			}
		}
		if remain {
			remainIds = append(remainIds, dbr)
		}
	}

	uwr.RoleIds = remainIds
	uwr.UpdateT = util.GetCurTime()

	// op history
	opAction := fmt.Sprintf("移除用户 roleIds %s", form.RoleIds)
	opHis := ophistory.NewOpHistory(curUser.Id, curUser.Name, opAction)
	uwr.History = append(uwr.History, opHis)

	err = userandrole.UpdateUserWithRole(db, uwr)
	middleware.StopExec(err)

	// 同步操作记录到用户历史记录
	_ = userapp.AppendOpHistoryToUser(db, form.UserId, opHis)
	returnfun.ReturnOKJson(c, uwr)
	return
}

// 读取指定 userid 的 roles 信息
func GetUserRolesHandler(c *gin.Context, ds *dbandmq.Ds) {
	id := c.Param("id")
	db := ds.CopyDs()
	defer db.Close()

	uwr, err := userandrole.GetUserRoles(db, id)
	middleware.StopExec(err)
	if uwr == nil {
		returnfun.ReturnErrJson(c, "无指定用户id的授权信息")
		return
	}

	returnfun.ReturnOKJson(c, uwr)
	return
}

// 读取授权了的用户列表
func QueryUWRHandler(c *gin.Context, ds *dbandmq.Ds) {
	var uwrs []*userandrole.UserWithRole
	page, size, skip := util.GetPageAndSize(c)

	query := bson.M{}

	db := ds.CopyDs()
	defer db.Close()

	Q := db.C(userandrole.CollectionNameUserWithRole).Find(query)
	total, err := Q.Count()
	middleware.StopExec(err)

	err = Q.Sort("-_id").Skip(skip).Limit(size).All(&uwrs)
	middleware.StopExec(err)

	retData := gin.H{
		"total": total,
		"page": page,
		"size": size,
		"data": uwrs,
	}
	returnfun.ReturnOKJson(c, retData)
	return
}

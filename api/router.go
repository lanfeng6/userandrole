package api

import (
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase/dbandmq"
	"github.com/leyle/userandrole/config"
)

// 角色路由
func RoleRouter(db *dbandmq.Ds, g *gin.RouterGroup) {
	auth := g.Group("", func(c *gin.Context) {
		Auth(c)
	})

	roleR := auth.Group("/role")

	// 明细
	itemR := roleR.Group("/item")
	{
		// 新建 item
		itemR.POST("", func(c *gin.Context) {
			CreateItemHandler(c, db)
		})

		// 修改 item
		itemR.PUT("/:id", func(c *gin.Context) {
			UpdateItemHandler(c, db)
		})

		// 删除 item
		itemR.DELETE("/:id", func(c *gin.Context) {
			DeleteItemHandler(c, db)
		})

		// 读取 item 明细
		itemR.GET("/:id", func(c *gin.Context) {
			GetItemInfoHandler(c, db)
		})

		// 搜索 item
		roleR.GET("/items", func(c *gin.Context) {
			QueryItemHandler(c, db)
		})
	}

	// 权限集合
	permissionR := roleR.Group("/permission")
	{
		// 新建 permission
		permissionR.POST("", func(c *gin.Context) {
			CreatePermissionHandler(c, db)
		})

		// 给权限添加 item，可多个
		permissionR.POST("/:id/additems", func(c *gin.Context) {
			AddItemToPermissionHandler(c, db)
		})

		// 给权限取消某个或某些 item，可多个
		permissionR.POST("/:id/delitems", func(c *gin.Context) {
			RemoveItemFromPermissionHandler(c, db)
		})

		// 修改权限基本信息
		permissionR.PUT("/:id", func(c *gin.Context) {
			UpdatePermissionHandler(c, db)
		})

		// 删除权限
		permissionR.DELETE("/:id", func(c *gin.Context) {
			DeletePermissionHandler(c, db)
		})

		// 读取权限明细
		permissionR.GET("/:id", func(c *gin.Context) {
			GetPermissionInfoHandler(c, db)
		})

		// 搜索权限列表
		roleR.GET("/permissions", func(c *gin.Context) {
			QueryPermissionHandler(c, db)
		})
	}

	// role
	rR := roleR.Group("/role")
	{
		// 新建 role
		rR.POST("", func(c *gin.Context) {
			CreateRoleHandler(c, db)
		})

		// 给 role 添加 permission
		rR.POST("/:id/addps", func(c *gin.Context) {
			AddPermissionsToRoleHandler(c, db)
		})

		// 从 role 中移除 permission
		rR.POST("/:id/delps", func(c *gin.Context) {
			RemovePermissionsFromRoleHandler(c, db)
		})

		// 修改 role 信息
		rR.PUT("/:id", func(c *gin.Context) {
			UpdateRoleInfoHandler(c, db)
		})

		// 删除role
		rR.DELETE("/:id", func(c *gin.Context) {
			DeleteRoleHandler(c, db)
		})

		// 查看 role 明细
		rR.GET("/:id", func(c *gin.Context) {
			GetRoleInfoHandler(c, db)
		})

		// 搜索 role
		roleR.GET("/roles", func(c *gin.Context) {
			QueryRoleHandler(c, db)
		})
	}
}

// 用户路由
func UserRouter(uo *UserOption, g *gin.RouterGroup) {
	auth := g.Group("", func(c *gin.Context) {
		Auth(c)
	})

	userR := auth.Group("/user")
	{
		// 新建一个账户密码，只能通过拥有相关权限的人来调用
		userR.POST("/idpasswd", func(c *gin.Context) {
			CreateLoginIdPasswdAccountHandler(c, uo)
		})

		// 修改密码,修改自己的密码
		userR.POST("/idpasswd/changepasswd", func(c *gin.Context) {
			UpdatePasswdHandler(c, uo)
		})

		// 已登录微信情况下，绑定手机号
		userR.POST("/wx/bindphone", func(c *gin.Context) {
			WeChatBindPhoneHandler(c, uo)
		})

		// 用户读取自己的信息
		userR.GET("/me", func(c *gin.Context) {
			MeHandler(c, uo)
		})

		// 退出登录
		userR.GET("/logout", func(c *gin.Context) {
			LogoutHandler(c, uo)
		})

		// 管理员创建一个手机号登录账户
		userR.POST("/phone", func(c *gin.Context) {
			CreateLoginPhoneHandler(c, uo)
		})

		// 管理员封禁用户
		userR.POST("/ban", func(c *gin.Context) {
			BanUserHandler(c, uo)
		})

		// 管理员解封用户
		userR.POST("/unban", func(c *gin.Context) {
			UnBanUserHandler(c, uo)
		})

		// 管理员重置用户密码
		userR.POST("/idpasswd/resetpasswd", func(c *gin.Context) {
			ResetPasswdHandler(c, uo)
		})

		// 管理员读取某个用户的详细信息
		userR.GET("/user/:id", func(c *gin.Context) {
			GetUserInfoHandler(c, uo)
		})

		// 查看用户的登录历史记录
		userR.GET("/loginhistory/:id", func(c *gin.Context) {
			GetUserLoginHistoryHandler(c, uo)
		})

		// 管理员搜索用户列表
		userR.GET("/users", func(c *gin.Context) {
			QueryUserHandler(c, uo)
		})
	}

	// 不需要 auth 的
	noAuthR := g.Group("/user")
	{
		// 账户密码登录
		noAuthR.POST("/idpasswd/login", func(c *gin.Context) {
			LoginByIdPasswdHandler(c, uo)
		})

		// 读取微信 appid
		noAuthR.GET("/wx/appid", func(c *gin.Context) {
			GetWeChatAppIdHandler(c, uo)
		})

		// 微信登录
		noAuthR.POST("/wx/login", func(c *gin.Context) {
			LoginByWeChatHandler(c, uo)
		})

		// 手机号注册、登录
		noAuthR.POST("/phone/sendsms", func(c *gin.Context) {
			SendSmsHandler(c, uo)
		})
		noAuthR.POST("/phone/checksms", func(c *gin.Context) {
			CheckSmsHandler(c, uo)
		})

		// token 验证
		noAuthR.POST("/token/check", func(c *gin.Context) {
			TokenCheckHandler(c, uo)
		})

	}
}

// 角色 <-> 用户 关联路由
func UserWithRoleRouter(db *dbandmq.Ds, g *gin.RouterGroup) {
	auth := g.Group("", func(c *gin.Context) {
		Auth(c)
	})

	uwrR := auth.Group("/uwr")
	{
		// 给 user 添加 roles
		uwrR.POST("/addroles", func(c *gin.Context) {
			AddRolesToUserHandler(c, db)
		})

		// 取消 user 的 roles
		uwrR.POST("/delroles", func(c *gin.Context) {
			RemoveRolesFromUserHandler(c, db)
		})

		// 读取指定用户的 roles 信息
		uwrR.GET("/user/:id", func(c *gin.Context) {
			GetUserRolesHandler(c, db)
		})

		// 搜索已经授权的用户列表
		uwrR.GET("/users", func(c *gin.Context) {
			QueryUWRHandler(c, db)
		})
	}
}

// 系统配置
func SystemConfRouter(conf *config.Config, g *gin.RouterGroup) {
	sysR := g.Group("/sys", func(c *gin.Context) {
		Auth(c)
	})
	{
		// 读取 redis 和 mongodb 配置
		sysR.GET("/conf", func(c *gin.Context) {
			GetMongodbAndRedisConfHandler(c, conf)
		})
	}
}
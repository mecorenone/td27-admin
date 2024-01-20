package authority

import (
	"fmt"
	"go.uber.org/zap"

	"server/global"
	modelAuthority "server/model/authority"
	authorityReq "server/model/authority/request"
	authorityRes "server/model/authority/response"
	modelBase "server/model/base"
	"server/model/common/request"
	"server/utils"
)

type UserService struct{}

func (us *UserService) GetUserInfo(userId uint) (userResults authorityRes.UserResult, err error) {
	err = global.TD27_DB.Table("base_user").Select("base_user.created_at,base_user.id,base_user.username,base_user.phone,base_user.email,base_user.active,base_user.role_model_id,authority_role.role_name").Joins("inner join authority_role on base_user.role_model_id = authority_role.id").Where("base_user.id = ?", userId).Scan(&userResults).Error
	return
}

// GetUsers 获取所有用户
func (us *UserService) GetUsers(pageInfo request.PageInfo) ([]authorityRes.UserResult, int64, error) {
	var userResults []authorityRes.UserResult
	var total int64

	db := global.TD27_DB.Model(&modelBase.UserModel{})

	// 分页
	err := db.Count(&total).Error
	if err != nil {
		return userResults, total, fmt.Errorf("分页count -> %v", err)
	} else {
		limit := pageInfo.PageSize
		offset := pageInfo.PageSize * (pageInfo.Page - 1)
		db = db.Limit(limit).Offset(offset)
		//err = db.Find(&list).Error
		// 左连接 查询出role_name
		db.Select("base_user.id,base_user.username,base_user.phone,base_user.email,base_user.active,base_user.role_model_id,authority_role.role_name").Joins("left join authority_role on base_user.role_model_id = authority_role.id").Scan(&userResults)
	}

	return userResults, total, err
}

// DeleteUser 删除用户
func (us *UserService) DeleteUser(id uint) (err error) {
	return global.TD27_DB.Where("id = ?", id).Unscoped().Delete(&modelBase.UserModel{}).Error
}

// AddUser 添加用户
func (us *UserService) AddUser(user authorityReq.AddUser) (err error) {
	err = global.TD27_DB.Where("id = ?", user.RoleModelID).First(&modelAuthority.RoleModel{}).Error
	if err != nil {
		global.TD27_LOG.Error("添加用户 -> 查询role", zap.Error(err))
		return err
	}

	var userModel modelBase.UserModel
	userModel.Username = user.Username
	userModel.Password = utils.MD5V([]byte(user.Password))
	userModel.Phone = user.Phone
	userModel.Email = user.Email
	userModel.Active = user.Active
	userModel.RoleModelID = user.RoleModelID

	return global.TD27_DB.Create(&userModel).Error
}

// EditUser 编辑用户
func (us *UserService) EditUser(user authorityReq.EditUser) (*authorityRes.UserResult, error) {
	var userModel modelBase.UserModel
	var userResult authorityRes.UserResult
	// 用户是否存在
	err := global.TD27_DB.Where("id = ?", user.Id).First(&userModel).Error
	if err != nil {
		global.TD27_LOG.Error("编辑用户 -> 查询Id", zap.Error(err))
		return nil, err
	}

	// 角色是否存在
	var roleModel modelAuthority.RoleModel
	err = global.TD27_DB.Where("id = ?", user.RoleModelID).First(&roleModel).Error
	if err != nil {
		global.TD27_LOG.Error("编辑用户 -> 查询role", zap.Error(err))
		return nil, err
	}

	updateV := make(map[string]interface{}, 5)
	updateV["username"] = user.Username
	updateV["active"] = user.Active
	updateV["role_model_id"] = user.RoleModelID
	updateV["phone"] = user.Phone
	updateV["email"] = user.Email

	err = global.TD27_DB.Model(&userModel).Updates(updateV).Error
	if err != nil {
		global.TD27_LOG.Error("编辑用户 -> update", zap.Error(err))
		return nil, err
	}

	userResult.ID = userModel.ID
	userResult.Username = userModel.Username
	userResult.Phone = userModel.Phone
	userResult.Email = userModel.Email
	userResult.Active = userModel.Active
	userResult.RoleName = roleModel.RoleName
	userResult.RoleModelID = userModel.RoleModelID

	return &userResult, nil
}

// ModifyPass 修改用户密码
func (us *UserService) ModifyPass(mp authorityReq.ModifyPass) (err error) {
	var userModel modelBase.UserModel
	err = global.TD27_DB.Where("id = ? and password = ?", mp.Id, utils.MD5V([]byte(mp.OldPassword))).First(&userModel).Error
	if err != nil {
		global.TD27_LOG.Error("修改用户密码 -> 查询用户", zap.Error(err))
		return err
	}
	return global.TD27_DB.Model(&userModel).Update("password", utils.MD5V([]byte(mp.NewPassword))).Error
}

// SwitchActive 切换启用状态
func (us *UserService) SwitchActive(sa authorityReq.SwitchActive) (err error) {
	var userModel modelBase.UserModel
	err = global.TD27_DB.Where("id = ?", sa.Id).First(&userModel).Error
	if err != nil {
		global.TD27_LOG.Error("切换启用状态 -> 查询用户", zap.Error(err))
		return err
	}
	if sa.Active {
		return global.TD27_DB.Model(&userModel).Update("active", true).Error
	} else {
		return global.TD27_DB.Model(&userModel).Update("active", false).Error
	}
}

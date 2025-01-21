package data

import (
	"context"
	"fmt"

	"codo-notice/internal/biz"
	"codo-notice/internal/impl/data/models"

	"github.com/ccheers/xpkg/generic/arrayx"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *Data
}

func NewIUserRepo(x *UserRepo) biz.IUserUseRepo {
	return x
}

func NewUserRepo(db *Data) *UserRepo {
	return &UserRepo{db: db}
}

func (x *UserRepo) List(ctx context.Context, query biz.UserQuery) ([]*biz.User, error) {
	db := x.db.DBWithContext(ctx)
	db, err := x.convertQuery(db, query)
	if err != nil {
		return nil, err
	}

	var result []*models.User
	if err := db.Model(&models.User{}).Find(&result).Error; err != nil {
		return nil, fmt.Errorf("%w: list user error", err)
	}

	return arrayx.Map(result, func(t *models.User) *biz.User {
		return x.convertDO(t)
	}), nil
}

func (x *UserRepo) Count(ctx context.Context, query biz.UserQuery) (uint32, error) {
	db := x.db.DBWithContext(ctx)
	db, err := x.convertQuery(db, query)
	if err != nil {
		return 0, err
	}

	var count int64
	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("%w: count user error", err)
	}
	return uint32(count), nil
}

func (x *UserRepo) Get(ctx context.Context, userID string) (*biz.User, error) {
	result := new(models.User)
	if err := x.db.DBWithContext(ctx).Model(&models.User{}).Where("user_id = ?", userID).First(result).Error; err != nil {
		return nil, fmt.Errorf("%w: get user by id error, userID=%s", err, userID)
	}
	return x.convertDO(result), nil
}

func (x *UserRepo) Save(ctx context.Context, user *biz.User) error {
	data := x.convertVO(user)
	var dst models.User
	result := x.db.DBWithContext(ctx).Model(&models.User{}).
		Where("user_id = ?", user.UserId).
		Attrs(data).
		FirstOrCreate(&dst)
	err := result.Error
	if err != nil {
		return fmt.Errorf("%w: save user error", err)
	}
	if result.RowsAffected == 1 {
		return nil
	}
	data.ID = dst.ID
	data.CreatedAt = dst.CreatedAt
	err = x.db.DBWithContext(ctx).Model(&models.User{}).Where("id = ?", dst.ID).Save(data).Error
	if err != nil {
		return fmt.Errorf("%w: save user error", err)
	}
	return nil
}

func (x *UserRepo) Delete(ctx context.Context, ids []uint32) error {
	return x.db.DBWithContext(ctx).Model(&models.User{}).Where("id in ?", ids).Unscoped().Delete(&models.User{}).Error
}

func (x *UserRepo) convertQuery(db *gorm.DB, query biz.UserQuery) (*gorm.DB, error) {
	reqTable := &ReqTable{
		PageSize:    int(query.PageSize),
		PageNum:     int(query.PageNum),
		Order:       query.Order,
		SearchText:  query.SearchText,
		SearchField: query.SearchField,
		Field:       query.Field,
		Cache:       query.Cache,
		FilterMap:   query.FilterMap,
	}
	condition, err := reqTable.convertQuery(new(models.User), []string{"data_source"}, nil)
	if err != nil {
		return nil, err
	}
	db = db.Where(condition.where, condition.values...).Order(condition.order)
	if len(query.UserIDNotIn) > 0 {
		db = db.Where("user_id not in ?", query.UserIDNotIn)
	}
	if query.ListAll {
		return db, nil
	}
	return db.Offset(condition.offset).Limit(condition.limit), nil
}

func (x *UserRepo) convertVO(user *biz.User) *models.User {
	data := &models.User{
		Username:   user.Username,
		Nickname:   user.Nickname,
		UserId:     user.UserId,
		DepId:      user.DepId,
		Dep:        user.Dep,
		Manager:    user.Manager,
		Avatar:     user.Avatar,
		Active:     user.Active,
		Tel:        user.Tel,
		Email:      user.Email,
		DataSource: user.DataSource,
		Disable:    user.Disable,
		DdID:       user.DdId,
		FsID:       user.FsId,
		WxID:       user.WxId,
	}
	data.ID = uint(user.ID)

	return data
}

func (x *UserRepo) convertDO(user *models.User) *biz.User {
	return &biz.User{
		ID:         uint32(user.ID),
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
		Username:   user.Username,
		Nickname:   user.Nickname,
		UserId:     user.UserId,
		DepId:      user.DepId,
		Dep:        user.Dep,
		Manager:    user.Manager,
		Avatar:     user.Avatar,
		Active:     user.Active,
		Tel:        user.Tel,
		Email:      user.Email,
		DataSource: user.DataSource,
		Disable:    user.Disable,
		DdId:       user.DdID,
		FsId:       user.FsID,
		WxId:       user.WxID,
	}
}

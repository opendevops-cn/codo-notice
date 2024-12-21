package data

import (
	"context"
	"encoding/json"

	"codo-notice/internal/biz"
	"codo-notice/internal/impl/data/models"

	"github.com/ccheers/xpkg/generic/arrayx"
	"gorm.io/gorm"
)

type RouterRepo struct {
	data *Data
}

func NewIRouterRepo(x *RouterRepo) biz.IRouterRepo {
	return x
}

func NewRouterRepo(data *Data) *RouterRepo {
	return &RouterRepo{data: data}
}

func (x *RouterRepo) List(ctx context.Context, query biz.RouterQuery) ([]*biz.Router, error) {
	db := x.data.DBWithContext(ctx)
	db, err := x.convertQuery(db, query)
	if err != nil {
		return nil, err
	}

	var result []*models.Router
	if err := db.Model(&models.Router{}).Find(&result).Error; err != nil {
		return nil, err
	}

	return arrayx.Map(result, func(t *models.Router) *biz.Router {
		return x.convertDO(t)
	}), nil
}

func (x *RouterRepo) Count(ctx context.Context, query biz.RouterQuery) (uint32, error) {
	db := x.data.DBWithContext(ctx)
	db, err := x.convertQuery(db, query)
	if err != nil {
		return 0, err
	}

	var count int64
	if err := db.Model(&models.Router{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return uint32(count), nil
}

func (x *RouterRepo) Get(ctx context.Context, id uint32) (*biz.Router, error) {
	result := new(models.Router)
	if err := x.data.DBWithContext(ctx).Model(&models.Router{}).Where("id = ?", id).First(result).Error; err != nil {
		return nil, err
	}
	return x.convertDO(result), nil
}

func (x *RouterRepo) Create(ctx context.Context, router *biz.Router) (*biz.Router, error) {
	data := x.convertVO(router)
	if err := x.data.DBWithContext(ctx).Create(data).Error; err != nil {
		return nil, err
	}
	return x.convertDO(data), nil
}

func (x *RouterRepo) Update(ctx context.Context, router *biz.Router, opts biz.RouterUpdateOptions) error {
	data := x.convertVO(router)
	if err := x.data.DBWithContext(ctx).Model(&models.Router{}).Where("id = ?", router.ID).Updates(data).Error; err != nil {
		return err
	}
	return nil
}

func (x *RouterRepo) Delete(ctx context.Context, ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}
	if err := x.data.DBWithContext(ctx).Where("id IN (?)", ids).Delete(&models.Router{}).Error; err != nil {
		return err
	}
	return nil
}

func (x *RouterRepo) convertQuery(db *gorm.DB, query biz.RouterQuery) (*gorm.DB, error) {
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
	condition, err := reqTable.convertQuery(new(models.Router), []string{"condition_list"}, nil)
	if err != nil {
		return nil, err
	}
	db = db.Where(condition.where, condition.values...).Order(condition.order)
	if query.ListAll {
		return db, nil
	}
	return db.Offset(condition.offset).Limit(condition.limit), nil
}

func (x *RouterRepo) convertVO(item *biz.Router) *models.Router {
	bs, _ := json.Marshal(item.ConditionList)
	data := &models.Router{
		CreatedBy:   item.CreatedBy,
		UpdatedBy:   item.UpdatedBy,
		Name:        item.Name,
		Description: item.Description,
		Status:      item.Status,
		ChannelId:   uint(item.ChannelID),
		Condition:   bs,
	}
	data.ID = uint(item.ID)

	return data
}

func (x *RouterRepo) convertDO(item *models.Router) *biz.Router {
	var conditionList biz.QueryGroup
	_ = json.Unmarshal(item.Condition, &conditionList)
	return &biz.Router{
		ID:            uint32(item.ID),
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
		CreatedBy:     item.CreatedBy,
		UpdatedBy:     item.UpdatedBy,
		Name:          item.Name,
		Description:   item.Description,
		Status:        item.Status,
		ChannelID:     uint32(item.ChannelId),
		ConditionList: conditionList,
	}
}

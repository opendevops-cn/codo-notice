package data

import (
	"context"
	"encoding/json"

	"codo-notice/internal/biz"
	"codo-notice/internal/impl/data/models"

	"github.com/ccheers/xpkg/generic/arrayx"
	"gorm.io/gorm"
)

type ChannelRepo struct {
	db *Data
}

func NewIChannelRepo(x *ChannelRepo) biz.IChannelRepo {
	return x
}

func NewChannelRepo(db *Data) *ChannelRepo {
	return &ChannelRepo{db: db}
}

func (x *ChannelRepo) List(ctx context.Context, query biz.ChannelQuery) ([]*biz.Channel, error) {
	db := x.db.DBWithContext(ctx)
	db, err := x.convertQuery(db, query)
	if err != nil {
		return nil, err
	}

	var result []*models.Channel
	if err := db.Model(&models.Channel{}).Find(&result).Error; err != nil {
		return nil, err
	}

	return arrayx.Map(result, func(t *models.Channel) *biz.Channel {
		return x.convertDO(t)
	}), nil
}

func (x *ChannelRepo) Count(ctx context.Context, query biz.ChannelQuery) (uint32, error) {
	db := x.db.DBWithContext(ctx)
	db, err := x.convertQuery(db, query)
	if err != nil {
		return 0, err
	}

	var count int64
	if err := db.Model(&models.Channel{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return uint32(count), nil
}

func (x *ChannelRepo) Get(ctx context.Context, id uint32) (*biz.Channel, error) {
	result := new(models.Channel)
	if err := x.db.DBWithContext(ctx).Model(&models.Channel{}).Where("id = ?", id).First(result).Error; err != nil {
		return nil, err
	}
	return x.convertDO(result), nil
}

func (x *ChannelRepo) Create(ctx context.Context, channel *biz.Channel) (*biz.Channel, error) {
	data := x.convertVO(channel)
	if err := x.db.DBWithContext(ctx).Model(&models.Channel{}).Create(data).Error; err != nil {
		return nil, err
	}
	return x.convertDO(data), nil
}

func (x *ChannelRepo) Update(ctx context.Context, channel *biz.Channel, opts biz.ChannelUpdateOptions) error {
	data := x.convertVO(channel)
	if err := x.db.DBWithContext(ctx).Model(&models.Channel{}).Where("id = ?", data.ID).Updates(data).Error; err != nil {
		return err
	}
	return nil
}

func (x *ChannelRepo) Delete(ctx context.Context, ids []uint32) error {
	if err := x.db.DBWithContext(ctx).Model(&models.Channel{}).Where("id IN ?", ids).Delete(&models.Channel{}).Error; err != nil {
		return err
	}
	return nil
}

func (x *ChannelRepo) convertQuery(db *gorm.DB, query biz.ChannelQuery) (*gorm.DB, error) {
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
	condition, err := reqTable.convertQuery(new(models.Channel), nil, nil)
	if err != nil {
		return nil, err
	}
	db = db.Where(condition.where, condition.values...).Order(condition.order)
	if query.ListAll {
		return db, nil
	}
	return db.Offset(condition.offset).Limit(condition.limit), nil
}

func (x *ChannelRepo) convertVO(item *biz.Channel) *models.Channel {
	contactPointsBS, _ := json.Marshal(item.ContactPoints)
	customItemsBS, _ := json.Marshal(item.CustomItems)
	data := &models.Channel{
		CreatedBy:     item.CreatedBy,
		UpdatedBy:     item.UpdatedBy,
		Name:          item.Name,
		Use:           item.Use,
		User:          item.User,
		Group:         arrayx.Map(item.Group, func(t uint32) uint { return uint(t) }),
		ContactPoints: contactPointsBS,
		CustomItems:   customItemsBS,
		DefaultRule:   item.DefaultRule,
	}
	data.ID = uint(item.ID)

	return data
}

func (x *ChannelRepo) convertDO(item *models.Channel) *biz.Channel {
	var contactPoints []*biz.ContactPoint
	_ = json.Unmarshal(item.ContactPoints, &contactPoints)
	var customItems []*biz.ContactPoint
	_ = json.Unmarshal(item.CustomItems, &customItems)
	return &biz.Channel{
		ID:            uint32(item.ID),
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
		CreatedBy:     item.CreatedBy,
		UpdatedBy:     item.UpdatedBy,
		Name:          item.Name,
		Use:           item.Use,
		User:          item.User,
		Group:         arrayx.Map(item.Group, func(t uint) uint32 { return uint32(t) }),
		ContactPoints: contactPoints,
		CustomItems:   customItems,
		DefaultRule:   item.DefaultRule,
	}
}

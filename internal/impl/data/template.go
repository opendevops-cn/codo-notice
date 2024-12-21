package data

import (
	"context"

	"codo-notice/internal/biz"
	"codo-notice/internal/impl/data/models"

	"github.com/ccheers/xpkg/generic/arrayx"
	"gorm.io/gorm"
)

type TemplateRepo struct {
	data *Data
}

func NewITemplateRepo(x *TemplateRepo) biz.ITemplateRepo {
	return x
}

func NewTemplateRepo(data *Data) *TemplateRepo {
	return &TemplateRepo{data: data}
}

func (x *TemplateRepo) List(ctx context.Context, query biz.TemplateQuery) ([]*biz.Template, error) {
	db := x.data.DBWithContext(ctx)
	db, err := x.convertQuery(db, query)
	if err != nil {
		return nil, err
	}

	var result []*models.Template
	if err := db.Model(&models.Template{}).Find(&result).Error; err != nil {
		return nil, err
	}

	return arrayx.Map(result, func(t *models.Template) *biz.Template {
		return x.convertDO(t)
	}), nil
}

func (x *TemplateRepo) Count(ctx context.Context, query biz.TemplateQuery) (uint32, error) {
	db := x.data.DBWithContext(ctx)
	db, err := x.convertQuery(db, query)
	if err != nil {
		return 0, err
	}

	var count int64
	if err := db.Model(&models.Template{}).Count(&count).Error; err != nil {
		return 0, err
	}

	return uint32(count), nil
}

func (x *TemplateRepo) Get(ctx context.Context, id uint32) (*biz.Template, error) {
	result := new(models.Template)
	if err := x.data.DBWithContext(ctx).Model(&models.Template{}).Where("id = ?", id).First(result).Error; err != nil {
		return nil, err
	}
	return x.convertDO(result), nil
}

func (x *TemplateRepo) Create(ctx context.Context, template *biz.Template) (*biz.Template, error) {
	data := x.convertVO(template)
	if err := x.data.DBWithContext(ctx).Model(&models.Template{}).Create(data).Error; err != nil {
		return nil, err
	}
	return x.convertDO(data), nil
}

func (x *TemplateRepo) Update(ctx context.Context, template *biz.Template, opts biz.TemplateUpdateOptions) error {
	data := x.convertVO(template)
	if err := x.data.DBWithContext(ctx).Model(&models.Template{}).
		Where("id = ?", template.ID).
		Updates(data).Error; err != nil {
		return err
	}
	return nil
}

func (x *TemplateRepo) Delete(ctx context.Context, ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}
	if err := x.data.DBWithContext(ctx).Model(&models.Template{}).
		Where("id IN (?)", ids).Delete(&models.Template{}).Error; err != nil {
		return err
	}
	return nil
}

func (x *TemplateRepo) convertQuery(db *gorm.DB, query biz.TemplateQuery) (*gorm.DB, error) {
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
	condition, err := reqTable.convertQuery(new(models.Template), nil, nil)
	if err != nil {
		return nil, err
	}
	db = db.Where(condition.where, condition.values...).Order(condition.order)
	if query.ListAll {
		return db, nil
	}
	return db.Offset(condition.offset).Limit(condition.limit), nil
}

func (x *TemplateRepo) convertVO(item *biz.Template) *models.Template {
	data := &models.Template{
		CreatedBy: item.CreatedBy,
		UpdatedBy: item.UpdatedBy,
		Name:      item.Name,
		Content:   item.Content,
		Type:      string(item.Type),
		Use:       item.Use,
		Default:   item.Default,
	}
	data.ID = uint(item.ID)

	return data
}

func (x *TemplateRepo) convertDO(item *models.Template) *biz.Template {
	return &biz.Template{
		ID:        uint32(item.ID),
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
		CreatedBy: item.CreatedBy,
		UpdatedBy: item.UpdatedBy,
		Name:      item.Name,
		Content:   item.Content,
		Type:      biz.NotifyType(item.Type),
		Use:       item.Use,
		Default:   item.Default,
		Path:      "",
	}
}

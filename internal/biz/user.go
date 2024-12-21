package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"codo-notice/internal/conf"

	"github.com/ccheers/xpkg/generic/arrayx"
	"github.com/ccheers/xpkg/sync/errgroup"
	"github.com/ccheers/xpkg/sync/try_lock"
	"github.com/opendevops-cn/codo-golang-sdk/client/xhttp"
	loggersdk "github.com/opendevops-cn/codo-golang-sdk/logger"
)

type IUserUseCase interface {
	List(ctx context.Context, query UserQuery) ([]*User, uint32, error)
	Get(ctx context.Context, id uint32) (*User, error)
}

type IUserUseRepo interface {
	List(ctx context.Context, query UserQuery) ([]*User, error)
	Count(ctx context.Context, query UserQuery) (uint32, error)
	Get(ctx context.Context, id uint32) (*User, error)
	Save(ctx context.Context, user *User) error
	Delete(ctx context.Context, ids []uint32) error
}

// User 用户信息
type User struct {
	// 用户ID，主键
	ID uint32
	// 创建时间
	CreatedAt time.Time
	// 更新时间
	UpdatedAt time.Time
	// 名称，最大长度128
	Username string
	// 昵称，最大长度128
	Nickname string
	// 用户标识，唯一且不为空，最大长度128
	UserId string
	// 部门标识，最大长度2048
	DepId string
	// 部门名称，最大长度2048
	Dep string
	// 管理者，最大长度128
	Manager string
	// 头像URL，最大长度1024
	Avatar string
	// 是否激活，默认为true
	Active bool
	// 手机号码，最大长度32
	Tel string
	// 邮箱地址，最大长度128
	Email string
	// 源数据，用于存储额外的动态数据
	DataSource json.RawMessage
	// 是否禁用，默认为false
	Disable bool
	// 钉钉ID，最大长度128
	DdId string
	// 飞书ID，最大长度128
	FsId string
}

type UserQuery struct {
	// 是否查询所有
	ListAll bool
	// 每页条数
	PageSize int32 `protobuf:"varint,1,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	// 第几页
	PageNum int32 `protobuf:"varint,2,opt,name=page_num,json=pageNum,proto3" json:"page_num,omitempty"`
	// 正序或倒序 ascend  descend
	Order string `protobuf:"bytes,3,opt,name=order,proto3" json:"order,omitempty"`
	// 全局搜索关键字
	SearchText string `protobuf:"bytes,4,opt,name=search_text,json=searchText,proto3" json:"search_text,omitempty"`
	// 搜索字段
	SearchField string `protobuf:"bytes,5,opt,name=search_field,json=searchField,proto3" json:"search_field,omitempty"`
	// 排序关键字
	Field string `protobuf:"bytes,6,opt,name=field,proto3" json:"field,omitempty"`
	// yes:缓存，no:不缓存
	Cache string `protobuf:"bytes,7,opt,name=cache,proto3" json:"cache,omitempty"`
	// 多字段搜索,精准匹配
	FilterMap map[string]interface{} `protobuf:"bytes,8,rep,name=filter_map,json=filterMap,proto3" json:"filter_map,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`

	// 不在用户ID列表中
	UserIDNotIn []string
}

type UserUseCase struct {
	logger *loggersdk.Helper
	repo   IUserUseRepo

	userSyncConf *conf.AppMetadata

	client xhttp.IClient

	cas try_lock.CASCommand
}

func NewIUserUseCase(x *UserUseCase) IUserUseCase {
	return x
}

func NewUserUseCase(ctx context.Context, logger loggersdk.Logger, bc *conf.Bootstrap, repo IUserUseRepo, client xhttp.IClient, cas try_lock.CASCommand) (*UserUseCase, func()) {
	x := &UserUseCase{
		logger:       loggersdk.NewHelper(logger),
		repo:         repo,
		userSyncConf: bc.Metadata,
		client:       client,
		cas:          cas,
	}

	ctx, cancel := context.WithCancel(ctx)
	eg := errgroup.WithCancel(ctx)
	eg.Go(func(ctx context.Context) error {
		for {
			err := x.syncUser(ctx)
			if err != nil {
				x.logger.Errorf(ctx, "[UserUseCase][syncUser]: %v", err)
			}
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(30 * time.Second):
			}
		}
	})

	return x, func() {
		cancel()
		_ = eg.Wait()
	}
}

func (x *UserUseCase) List(ctx context.Context, query UserQuery) ([]*User, uint32, error) {
	list, err := x.repo.List(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	query.ListAll = true
	count, err := x.repo.Count(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	return list, count, nil
}

func (x *UserUseCase) Get(ctx context.Context, id uint32) (*User, error) {
	return x.repo.Get(ctx, id)
}

func (x *UserUseCase) syncUser(ctx context.Context) error {
	type UCReply struct {
		Code  int    `json:"code"`
		Msg   string `json:"msg"`
		Count int    `json:"count"`
		Data  []struct {
			Id              int      `json:"id"`
			Username        string   `json:"username"`
			Nickname        string   `json:"nickname"`
			Email           string   `json:"email"`
			Tel             string   `json:"tel"`
			Department      string   `json:"department"`
			Superuser       string   `json:"superuser"`
			Avatar          string   `json:"avatar"`
			Source          string   `json:"source"`
			SourceAccountId string   `json:"source_account_id"`
			Manager         string   `json:"manager"`
			DdId            string   `json:"dd_id"`
			Status          string   `json:"status"`
			HaveToken       string   `json:"have_token"`
			FsOpenId        string   `json:"fs_open_id"`
			FsId            string   `json:"fs_id"`
			ExtInfo         struct{} `json:"ext_info"`
			LastIp          string   `json:"last_ip"`
			LastLogin       string   `json:"last_login"`
			CreateTime      string   `json:"create_time"`
			UpdateTime      string   `json:"update_time"`
		} `json:"data"`
	}
	const (
		lockKey = "codo:notice:sync_user"
	)

	// 抢占失败，直接返回
	cancel, err := try_lock.SimpleDistributedTryLock(x.cas, lockKey, time.Second)
	if err != nil {
		return nil
	}
	defer cancel()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/p/v4/user/", x.userSyncConf.GatewayPrefix), nil)
	if err != nil {
		return err
	}
	req.Header.Set("auth-key", x.userSyncConf.GatewayToken)
	req = req.WithContext(ctx)
	resp, err := x.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("[UserUseCase][syncUser][client.Do]: %w", err)
	}
	defer resp.Body.Close()
	var dst UCReply
	err = json.NewDecoder(resp.Body).Decode(&dst)
	if err != nil {
		return fmt.Errorf("[UserUseCase][syncUser][json.Decode]: %w", err)
	}
	allIDs := make([]string, 0, len(dst.Data))
	for _, datum := range dst.Data {
		allIDs = append(allIDs, strconv.Itoa(datum.Id))
		// insert
		_ = x.repo.Save(ctx, &User{
			Username:   datum.Username,
			Nickname:   datum.Nickname,
			UserId:     strconv.Itoa(datum.Id),
			DepId:      "",
			Dep:        "",
			Manager:    "",
			Avatar:     "",
			Active:     true,
			Tel:        datum.Tel,
			Email:      datum.Email,
			DataSource: nil,
			Disable:    false,
			DdId:       "",
			FsId:       datum.FsId,
		})
	}

	query := UserQuery{
		ListAll:     true,
		UserIDNotIn: allIDs,
	}
	users, err := x.repo.List(ctx, query)
	if err != nil {
		return fmt.Errorf("[UserUseCase][syncUser][repo.List]: %w, query=%+v", err, query)
	}
	if len(users) == 0 {
		return nil
	}

	// 删除不存在的用户
	err = x.repo.Delete(ctx, arrayx.Map(users, func(t *User) uint32 {
		return t.ID
	}))
	if err != nil {
		return fmt.Errorf("[UserUseCase][syncUser][repo.Delete]: %w, users=%+v", err, users)
	}
	return nil
}

package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"codo-notice/internal/impl/alerts"

	"github.com/go-redis/redis/v8"
)

type LarkCardCallbackRepo struct {
	client *redis.Client
}

func NewILarkCardCallbackRepo(x *LarkCardCallbackRepo) alerts.ILarkCardCallbackRepo {
	return x
}

func NewLarkCardCallbackRepo(client *redis.Client) *LarkCardCallbackRepo {
	return &LarkCardCallbackRepo{client: client}
}

func (x *LarkCardCallbackRepo) Create(ctx context.Context, info *alerts.LarkCardCallbackInfo) error {
	key := fmt.Sprintf("codo-notice:lark_card_callback:%s", info.WebhookUUID)
	bs, _ := json.Marshal(info)
	err := x.client.Set(ctx, key, bs, time.Hour*24*7).Err()
	if err != nil {
		return fmt.Errorf("failed to create lark card callback: %w", err)
	}
	return nil
}

func (x *LarkCardCallbackRepo) Get(ctx context.Context, webhookUUID string) (*alerts.LarkCardCallbackInfo, error) {
	key := fmt.Sprintf("codo-notice:lark_card_callback:%s", webhookUUID)

	bs, err := x.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get lark card callback: %w", err)
	}
	var info alerts.LarkCardCallbackInfo
	err = json.Unmarshal(bs, &info)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal lark card callback: %w", err)
	}
	return &info, nil
}

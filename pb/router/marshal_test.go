package router

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Ccheers/protoc-gen-go-kratos-http/khttp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMarshal(t *testing.T) {
	data, _ := json.Marshal(&RouterDTO{
		Id:            1,
		CreatedAt:     timestamppb.New(time.Now()),
		UpdatedAt:     timestamppb.New(time.Now()),
		CreatedBy:     "1",
		UpdatedBy:     "2",
		Name:          "3",
		Description:   "4",
		Status:        "5",
		ChannelId:     6,
		ConditionList: khttp.NewRawJSON([]byte(`[["1","2","3","4","5","6"],["1","2","3","4","5","6"],["1","2","3","4","5","6"]]`)),
	})

	t.Log(string(data))
}

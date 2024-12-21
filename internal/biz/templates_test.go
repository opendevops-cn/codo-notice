package biz

import (
	"testing"
	"time"
)

func TestTemplate_NotifyTemplatePath(t *testing.T) {
	type fields struct {
		ID        uint32
		CreatedAt time.Time
		UpdatedAt time.Time
		CreatedBy string
		UpdatedBy string
		Name      string
		Content   string
		Type      NotifyType
		Use       string
		Default   string
		Path      string
	}
	type args struct {
		gatewayPrefix string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "1",
			fields: fields{
				ID:        0,
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
				CreatedBy: "",
				UpdatedBy: "",
				Name:      "test",
				Content:   "",
				Type:      NotifyQiYeWXApp,
				Use:       "",
				Default:   "",
				Path:      "",
			},
			args: args{
				gatewayPrefix: "http://127.0.0.1:8000/",
			},
			want: "http://127.0.0.1:8000/api/noc/v1/alert?at=域账号&tpl=test&type=wxapp",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := &Template{
				ID:        tt.fields.ID,
				CreatedAt: tt.fields.CreatedAt,
				UpdatedAt: tt.fields.UpdatedAt,
				CreatedBy: tt.fields.CreatedBy,
				UpdatedBy: tt.fields.UpdatedBy,
				Name:      tt.fields.Name,
				Content:   tt.fields.Content,
				Type:      tt.fields.Type,
				Use:       tt.fields.Use,
				Default:   tt.fields.Default,
				Path:      tt.fields.Path,
			}
			if got := x.NotifyTemplatePath(tt.args.gatewayPrefix); got != tt.want {
				t.Errorf("NotifyTemplatePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

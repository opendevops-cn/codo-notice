package alerts

import (
	"bytes"
	"context"
	tmplhtml "html/template"
	"net"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
)

type TemplateRender struct{}

func NewTemplateRender() *TemplateRender {
	return &TemplateRender{}
}

func (x *TemplateRender) Render(ctx context.Context, templateContent string, data map[string]string) (string, error) {
	tmpl := x.getTmpl(ctx)
	tmpl, err := tmpl.Parse(templateContent)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (x *TemplateRender) getTmpl(ctx context.Context) *template.Template {
	tmpl := template.New("tmpl")
	funcMap := sprig.TxtFuncMap()
	funcMap["match"] = regexp.MatchString
	funcMap["reReplaceAll"] = func(pattern, repl, text string) string {
		return regexp.MustCompile(pattern).ReplaceAllString(text, repl)
	}
	funcMap["safeHtml"] = func(text string) tmplhtml.HTML { return tmplhtml.HTML(text) }
	funcMap["nowLocal"] = x.nowLocal
	funcMap["nowUtc"] = x.nowUtc
	funcMap["local2Utc"] = x.local2Utc
	funcMap["utc2Local"] = x.utc2Local
	funcMap["timeFormat"] = x.timeFormat2
	for k, f := range x.expandFunc() {
		funcMap[k] = f
	}
	tmpl = tmpl.Funcs(funcMap)
	return tmpl
}

func (x *TemplateRender) expandFunc() template.FuncMap {
	return template.FuncMap{
		"stripPort": func(hostPort string) string {
			host, _, err := net.SplitHostPort(hostPort)
			if err != nil {
				return hostPort
			}
			return host
		},
		"stripDomain": func(hostPort string) string {
			host, port, err := net.SplitHostPort(hostPort)
			if err != nil {
				host = hostPort
			}
			ip := net.ParseIP(host)
			if ip != nil {
				return hostPort
			}
			host = strings.Split(host, ".")[0]
			if port != "" {
				return net.JoinHostPort(host, port)
			}
			return host
		},
	}
}

const (
	DefaultLayout = "2006-01-02 15:04:05"
	RFC3339Layout = "2006-01-02T15:04:05Z"
)

func (x *TemplateRender) nowLocal() string {
	return time.Now().Format(DefaultLayout)
}

func (x *TemplateRender) nowUtc() string {
	return time.Now().UTC().Format(RFC3339Layout)
}

// utc2Local
// 将utc时间字符串转成本地时间字符串
// utcStr: 2021-06-16T10:20:45Z
func (x *TemplateRender) utc2Local(utcStr string) string {
	t, err := time.ParseInLocation(RFC3339Layout, utcStr, time.UTC)
	if err != nil {
		return ""
	}
	return x.timeFormat(t.Local())
}

// timeFormat
// 默认时间格式化
func (x *TemplateRender) timeFormat(t time.Time) string {
	return t.Format(DefaultLayout)
}

// timeUtcFormat
// 将指定时间转成utc时间格式化
func (x *TemplateRender) timeUtcFormat(t time.Time) string {
	return t.UTC().Format(RFC3339Layout)
}

// local2Utc
// 将本地时间字符串转成utc时间字符串
// localStr: 2021-06-16 18:20:45
func (x *TemplateRender) local2Utc(localStr string) string {
	t, err := time.ParseInLocation(DefaultLayout, localStr, time.Local)
	if err != nil {
		return ""
	}
	return x.timeUtcFormat(t)
}

// timeFormat2 解析时间
func (x *TemplateRender) timeFormat2(timeStr string, timeFormat ...string) string {
	format := DefaultLayout
	if len(timeFormat) > 0 {
		format = timeFormat[0]
	}
	result, err := time.Parse("2006-01-02T15:04:05.999999999Z", timeStr)
	if err != nil {
		result, err = time.Parse("2006-01-02T15:04:05.999999999+08:00", timeStr)
	}
	if err != nil {
		result, err = time.Parse("2006-01-02 15:04:05", timeStr)
	}
	if err != nil {
		return timeStr
	}
	return result.Format(format)
}

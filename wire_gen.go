// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"codo-notice/internal/biz"
	"codo-notice/internal/conf"
	"codo-notice/internal/dep"
	"codo-notice/internal/imiddleware"
	"codo-notice/internal/impl/alerts"
	"codo-notice/internal/impl/data"
	"codo-notice/internal/server"
	"codo-notice/internal/service"
	"context"

	"github.com/go-kratos/kratos/v2"
	"github.com/opendevops-cn/codo-golang-sdk/logger"

	_ "go.uber.org/automaxprocs"
)

// Injectors from wire.go:

// wireApp init kratos application.
func wireApp(contextContext context.Context, loggerLogger logger.Logger, bootstrap *conf.Bootstrap) (*kratos.App, func(), error) {
	logLogger, err := dep.NewLogger(bootstrap, loggerLogger)
	if err != nil {
		return nil, nil, err
	}
	meterProvider, err := dep.NewMeterProvider(bootstrap)
	if err != nil {
		return nil, nil, err
	}
	textMapPropagator := dep.NewTextMapPropagator()
	tracerProvider, err := dep.NewTracerProvider(contextContext, bootstrap, textMapPropagator, logLogger)
	if err != nil {
		return nil, nil, err
	}
	db, cleanup, err := dep.NewMysql(bootstrap, logLogger, tracerProvider)
	if err != nil {
		return nil, nil, err
	}
	client, cleanup2, err := dep.NewRedis(contextContext, logLogger, bootstrap, tracerProvider)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	dataData, err := data.NewData(db, client)
	if err != nil {
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	channelRepo := data.NewChannelRepo(dataData)
	iChannelRepo := data.NewIChannelRepo(channelRepo)
	channelUseCase := biz.NewChannelUseCase(iChannelRepo)
	iChannelUseCase := biz.NewIChannelUseCase(channelUseCase)
	channelService := service.NewChannelService(iChannelUseCase)
	routerRepo := data.NewRouterRepo(dataData)
	iRouterRepo := data.NewIRouterRepo(routerRepo)
	routerUseCase := biz.NewRouterUseCase(iRouterRepo)
	iRouterUseCase := biz.NewIRouterUseCase(routerUseCase)
	userRepo := data.NewUserRepo(dataData)
	iUserUseRepo := data.NewIUserRepo(userRepo)
	iClient, err := dep.NewXHTTPClient()
	if err != nil {
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	casCommand := dep.NewRedisCas(client)
	userUseCase, cleanup3 := biz.NewUserUseCase(contextContext, loggerLogger, bootstrap, iUserUseRepo, iClient, casCommand)
	iUserUseCase := biz.NewIUserUseCase(userUseCase)
	templateRender := alerts.NewTemplateRender()
	aliyunDianhuaAlerter := alerts.NewAliyunDianhuaAlerter(templateRender)
	aliyunDuanxinAlerter := alerts.NewAliyunDuanxinAlerter(templateRender)
	dingTalkWebhookAlerter := alerts.NewDingTalkWebhookAlerter(templateRender, iClient)
	dingTalkAppAlerter := alerts.NewDingTalkAppAlerter(templateRender, iClient)
	emailAlerter := alerts.NewEmailAlerter(templateRender)
	feishuWebhookAlerter := alerts.NewFeishuWebhookAlerter(templateRender, iClient)
	webhookAlerter := alerts.NewWebhookAlerter(templateRender, iClient)
	qiYeWeiXinAlerter := alerts.NewQiYeWeiXinAlerter(iClient, templateRender)
	qiYeWeiXinAppAlerter := alerts.NewQiYeWeiXinAppAlerter(iClient, templateRender)
	tencentCloudDHAlerter := alerts.NewTencentCloudDHAlerter(templateRender)
	tencentCloudDXAlerter := alerts.NewTencentCloudDXAlerter(templateRender)
	larkCardCallbackRepo := data.NewLarkCardCallbackRepo(client)
	iLarkCardCallbackRepo := data.NewILarkCardCallbackRepo(larkCardCallbackRepo)
	iMsgBus := dep.NewMsgbus(client)
	iSharedStorage := dep.NewSharedStorage(client)
	iTopicManager := dep.NewTopicManager(contextContext, iMsgBus, casCommand, iSharedStorage)
	otelOptions := dep.NewOtelOptions()
	feishuAppAlerter, cleanup4 := alerts.NewFeishuAppAlerter(contextContext, loggerLogger, iLarkCardCallbackRepo, templateRender, iClient, bootstrap, iMsgBus, iTopicManager, otelOptions)
	iAlerterList := alerts.NewAlerterList(aliyunDianhuaAlerter, aliyunDuanxinAlerter, dingTalkWebhookAlerter, dingTalkAppAlerter, emailAlerter, feishuWebhookAlerter, webhookAlerter, qiYeWeiXinAlerter, qiYeWeiXinAppAlerter, tencentCloudDHAlerter, tencentCloudDXAlerter, feishuAppAlerter)
	templateRepo := data.NewTemplateRepo(dataData)
	iTemplateRepo := data.NewITemplateRepo(templateRepo)
	templateUseCase := biz.NewTemplateUseCase(iTemplateRepo, bootstrap)
	iTemplateUseCase := biz.NewITemplateUseCase(templateUseCase)
	alertUseCase, cleanup5 := biz.NewAlertUseCase(iUserUseCase, bootstrap, iAlerterList, loggerLogger, iRouterUseCase, iChannelUseCase, iTemplateUseCase)
	iAlertUseCase := biz.NewIAlertUseCase(alertUseCase)
	routerService := service.NewRouterService(iRouterUseCase, iAlertUseCase, iUserUseCase)
	templateService := service.NewTemplateService(iTemplateUseCase, iUserUseCase, iAlertUseCase)
	userService := service.NewUserService(iUserUseCase)
	jwtMiddleware := imiddleware.NewJWTMiddleware(bootstrap)
	httpServer, err := server.NewHTTPServer(bootstrap, logLogger, meterProvider, tracerProvider, channelService, routerService, templateService, userService, jwtMiddleware)
	if err != nil {
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	pprofServer, err := server.NewPprofServer(bootstrap)
	if err != nil {
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	prometheusServer, err := server.NewPrometheusServer(bootstrap)
	if err != nil {
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	registrar, err := dep.NewRegister(contextContext, logLogger, bootstrap)
	if err != nil {
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	hookUseCase := biz.NewHookUseCase(bootstrap, iMsgBus, iTopicManager, otelOptions)
	iHookUseCase := biz.NewIHookUseCase(hookUseCase)
	hookService := service.NewHookService(iHookUseCase)
	thirdPartHookServer, err := server.NewThirdPartHookServer(bootstrap, logLogger, meterProvider, tracerProvider, hookService)
	if err != nil {
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	app, err := newApp(bootstrap, logLogger, httpServer, pprofServer, prometheusServer, registrar, thirdPartHookServer)
	if err != nil {
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	return app, func() {
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
	}, nil
}
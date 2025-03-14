package main

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/experimental/concurrent"
	"github.com/sasaki77/archiverappliance-datasource/pkg/aalive"
	"github.com/sasaki77/archiverappliance-datasource/pkg/archiverappliance"
	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
)

type ArchiverDatasource struct {
	// Structure defined by grafana-plugin-sdk-go. Implements QueryData and CheckHealth.
	//im instancemgmt.InstanceManager
	config models.DatasourceSettings
	client archiverappliance.Client
}

func newArchiverDataSource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	config, err := models.LoadSettings(settings)
	if err != nil {
		return nil, err
	}

	client, err := archiverappliance.NewAAClient(ctx, config.URL)
	if err != nil {
		return nil, err
	}

	return &ArchiverDatasource{config: config, client: client}, nil
}

func (td *ArchiverDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return concurrent.QueryData(ctx, req, td.handleSingleQueryData, 10)
}

func (td *ArchiverDatasource) handleSingleQueryData(ctx context.Context, q concurrent.Query) (res backend.DataResponse) {
	return archiverappliance.Query(ctx, q.DataQuery, td.client, td.config)
}

func (td *ArchiverDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	// the frontend health check is still used at this time
	var status = backend.HealthStatusOk
	var message = "This is a fake success message"

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

func (td *ArchiverDatasource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	log.DefaultLogger.Debug("SubscribeStream called", "request", req)

	status := backend.SubscribeStreamStatusPermissionDenied
	// Allow subscribing only on expected path.
	if aalive.IsPathValid(req.Path) {
		status = backend.SubscribeStreamStatusOK
	}
	return &backend.SubscribeStreamResponse{
		Status: status,
	}, nil
}

func (td *ArchiverDatasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	log.DefaultLogger.Debug("RunStream called", "request", req)

	pvname := aalive.ConvURL2PV(req.Path)

	wsDataProxy, err := aalive.NewWsDataProxy(ctx, sender, pvname, td.config.LiveUpdateURI)
	if err != nil {
		errCtx := "Starting WebSocket"
		log.DefaultLogger.Error(errCtx, "error", err.Error())
		aalive.SendErrorFrame(fmt.Sprintf("%s: %s", errCtx, err.Error()), sender)

		return err
	}

	go wsDataProxy.ProxyMessage()
	go wsDataProxy.ReadMessage()

	select {
	case <-ctx.Done():
		wsDataProxy.Done <- true
		log.DefaultLogger.Debug("Closing Channel", "channel", req.Path)
		return nil
	case rError := <-wsDataProxy.ReadingErrors:
		log.DefaultLogger.Error("Error reading the websocket", "error", err)
		aalive.SendErrorFrame(fmt.Sprintf("%s: %s", "Error reading the websocket", err), sender)

		log.DefaultLogger.Debug("Closing Channel due an error to read websocket", "channel", req.Path)

		return rError
	}
}

func (td *ArchiverDatasource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Debug("PublishStream called", "request", req)

	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

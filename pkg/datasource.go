package main

import (
	"context"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/sasaki77/archiverappliance-datasource/pkg/aalive"
	"github.com/sasaki77/archiverappliance-datasource/pkg/archiverappliance"
	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
)

type ArchiverDatasource struct {
	// Structure defined by grafana-plugin-sdk-go. Implements QueryData and CheckHealth.
	//im instancemgmt.InstanceManager
}

func newArchiverDataSource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &ArchiverDatasource{}, nil
}

func (td *ArchiverDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// Structure defined by grafana-plugin-sdk-go. QueryData should unpack the req argument into individual queries.

	config, err := models.LoadSettings(req.PluginContext)
	if err != nil {
		return nil, err
	}

	client, err := archiverappliance.NewAAClient(ctx, config.URL)
	if err != nil {
		return nil, err
	}

	response := archiverappliance.Query(ctx, client, req, config)

	return response, nil
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
	log.DefaultLogger.Info("SubscribeStream called", "request", req)

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
	log.DefaultLogger.Info("RunStream called", "request", req)

	// Create the same data frame as for query data.
	pvname := aalive.ConvURL2PV(req.Path)
	frame := data.NewFrame(pvname)

	// Add fields (matching the same schema used in QueryData).
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, make([]time.Time, 1)),
		data.NewField(pvname, nil, make([]int64, 1)),
	)

	counter := 0

	// Stream data frames periodically till stream closed by Grafana.
	for {
		select {
		case <-ctx.Done():
			log.DefaultLogger.Info("Context done, finish streaming", "path", req.Path)
			return nil
		case <-time.After(time.Second):
			// Send new data periodically.
			frame.Fields[0].Set(0, time.Now())
			frame.Fields[1].Set(0, int64(10*(counter%2+1)))

			counter++

			err := sender.SendFrame(frame, data.IncludeAll)
			if err != nil {
				log.DefaultLogger.Error("Error sending frame", "error", err)
				continue
			}
		}
	}
}

func (td *ArchiverDatasource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Info("PublishStream called", "request", req)

	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

package main

import (
	"context"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
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

	client, err := archiverappliance.NewAAClient(ctx, config)
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

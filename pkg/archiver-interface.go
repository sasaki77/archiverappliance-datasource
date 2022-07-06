package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type ArchiverDatasource struct {
	// Structure defined by grafana-plugin-sdk-go. Implements QueryData and CheckHealth.
	im instancemgmt.InstanceManager
}

type QueryMgr struct {
	Res    backend.DataResponse
	QRefID string
}

func newArchiverDataSource() datasource.ServeOpts {
	// Create a new instance manager
	log.DefaultLogger.Debug("Starting newArchiverDataSource")

	im := datasource.NewInstanceManager(newArchiverDataSourceInstance)
	ds := &ArchiverDatasource{
		im: im,
	}

	return datasource.ServeOpts{
		QueryDataHandler:   ds,
		CheckHealthHandler: ds,
	}
}

func (td *ArchiverDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// Structure defined by grafana-plugin-sdk-go. QueryData should unpack the req argument into individual queries.

	config, err := LoadSettings(req.PluginContext)
	if err != nil {
		return nil, err
	}

	// create response struct
	response := backend.NewQueryDataResponse()
	responsePipe := make(chan QueryMgr)

	for _, q := range req.Queries {
		go func(ctx context.Context, q backend.DataQuery, config DatasourceSettings, responsePipe chan QueryMgr) {
			res := backend.DataResponse{}
			qm, err := ReadQueryModel(q)

			if err != nil {
				res.Error = err
			} else {
				res = Query(ctx, qm, config)
			}

			responsePipe <- QueryMgr{
				Res:    res,
				QRefID: q.RefID,
			}
		}(ctx, q, config, responsePipe)
	}

	timeoutDurationSeconds := 30 // units are seconds
	timeoutDuration, _ := time.ParseDuration(strconv.Itoa(timeoutDurationSeconds) + "s")
	timeoutPipe := time.After(timeoutDuration)

queryCollector:
	for range req.Queries {
		// save the response in a hashmap
		// based on with RefID as identifier
		select {
		case rtn := <-responsePipe:
			response.Responses[rtn.QRefID] = rtn.Res
		case <-timeoutPipe:
			log.DefaultLogger.Warn("Timeout limit for QueryData has been reached")
			break queryCollector
		}
	}
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

type archiverInstanceSettings struct {
	httpClient *http.Client
}

func newArchiverDataSourceInstance(setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	// Adheres to structure defined by grafana-plugin-sdk-go
	return &archiverInstanceSettings{
		httpClient: &http.Client{},
	}, nil
}

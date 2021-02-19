package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

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

type ArchiverDatasource struct {
    // Structure defined by grafana-plugin-sdk-go. Implements QueryData and CheckHealth.
    im instancemgmt.InstanceManager
}

func (td *ArchiverDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
    // Structure defined by grafana-plugin-sdk-go. QueryData should unpack the req argument into individual queries.

    // create response struct
    response := backend.NewQueryDataResponse()
    // IMPLEMENT HERE
    for _, q := range req.Queries {

        res := td.query(ctx, q, req.PluginContext)

        // save the response in a hashmap
        // based on with RefID as identifier
        response.Responses[q.RefID] = res
    }

    return response, nil
}

func (td *ArchiverDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
    var status = backend.HealthStatusOk
    var message = "This is a fake success message"

    return &backend.CheckHealthResult{
        Status:     status,
        Message:    message,
    }, nil
}

func (td *ArchiverDatasource) query(ctx context.Context, query backend.DataQuery, pluginctx backend.PluginContext) backend.DataResponse {

    // Unmarshal the json into our queryModel
    var qm ArchiverQueryModel

    response := backend.DataResponse{}

    response.Error = json.Unmarshal(query.JSON, &qm)
    if response.Error != nil {
        return response
    }

    // make the query and compile the results into a SingleData instance
    responseData := make([]SingleData, 0)
    targetPvList := make([]string,0)
    if qm.Regex {
        // If the user is using a regex to specify the PVs, parse and resolve the regex expression first

        // assemble the list of PVs to be queried for
        regexUrl := BuildRegexUrl(qm.Target, pluginctx)
        regexQueryResponse, _ := ArchiverRegexQuery(regexUrl)
        targetPvList, _ = ArchiverRegexQueryParser(regexQueryResponse)
    } else {
        // If a regex is not being used, only check for listed PVs
        targetPvList = IsolateBasicQuery(qm.Target)
    }

    // execute the individual queries
    for _, targetPv := range targetPvList {
        parsedResponse, _ := ExecuteSingleQuery(targetPv, query, pluginctx, qm)
        responseData = append(responseData, parsedResponse)
    }

    // Apply Functions to the data
    var funcErr error
    responseData, funcErr = ApplyFunctions(responseData, qm)
    if funcErr != nil {
        log.DefaultLogger.Warn("Error applying functions")
    }


    // for each query response, compile the data into response.Framse
    for _, singleResponse := range responseData {
        // create data frame response
        frame := data.NewFrame("response")

        //add the time dimension
        frame.Fields = append(frame.Fields,
            data.NewField("time", nil, singleResponse.Times),
        )

        // add values 
        frame.Fields = append(frame.Fields, 
            data.NewField("values", nil, singleResponse.Values),
        )

        // add the frames to the response
        response.Frames = append(response.Frames, frame)
    }

    return response
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


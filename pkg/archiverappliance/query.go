package archiverappliance

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/live"
	"github.com/sasaki77/archiverappliance-datasource/pkg/aalive"
	"github.com/sasaki77/archiverappliance-datasource/pkg/functions"
	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
)

func Query(ctx context.Context, q backend.DataQuery, c Client, config models.DatasourceSettings) backend.DataResponse {
	res := backend.DataResponse{}
	qm, err := models.ReadQueryModel(q, config)

	if err != nil {
		res.Error = err
		return res
	}

	res = singleQuery(ctx, qm, c, config)

	return res
}

type queryResponse struct {
	response models.SingleData
	err      error
}

func singleQuery(_ context.Context, qm models.ArchiverQueryModel, client Client, config models.DatasourceSettings) backend.DataResponse {

	targetPvList := makeTargetPVList(client, qm.Target, qm.Regex, qm.MaxNumPVs)

	// execute the individual queries
	responseData := make([]*models.SingleData, 0, len(targetPvList))
	responsePipe := make(chan queryResponse)

	// Create timeout. If any request routines take longer than timeoutDurationSeconds to execute, they will be dropped.
	timeoutDurationSeconds := 30 // units are seconds
	timeoutDuration, _ := time.ParseDuration(strconv.Itoa(timeoutDurationSeconds) + "s")
	timeoutPipe := time.After(timeoutDuration)

	// create goroutines for individual requests
	for _, targetPv := range targetPvList {
		go func(targetPv string, pipe chan queryResponse) {
			parsedResponse, err := client.ExecuteSingleQuery(targetPv, qm)
			pipe <- queryResponse{response: parsedResponse, err: err}
		}(targetPv, responsePipe)
	}

	// Collect responses from the request goroutines
	var responseErr error
responseCollector:
	for range targetPvList {
		select {
		case response := <-responsePipe:
			if response.err != nil {
				if qm.IgnoreEmptyErr && errors.Is(response.err, errEmptyResponse) {
					continue
				}

				if responseErr == nil {
					responseErr = response.err
				}

				continue
			}
			responseData = append(responseData, &response.response)
		case <-timeoutPipe:
			log.DefaultLogger.Warn("Timeout limit for query has been reached")
			break responseCollector
		}
	}

	// Apply Alias to the data
	var aliasErr error
	responseData, aliasErr = applyAlias(responseData, qm)
	if aliasErr != nil {
		log.DefaultLogger.Warn("Error applying alias")
	}

	sort.Slice(responseData, func(i, j int) bool { return responseData[i].Name < responseData[j].Name })

	// Apply Functions to the data
	var funcErr error
	responseData, funcErr = functions.ApplyFunctions(responseData, qm)
	if funcErr != nil {
		log.DefaultLogger.Warn("Error applying functions")
	}

	// Extrapolate data as necessary
	for idx, data := range responseData {
		responseData[idx] = dataExtrapol(data, qm)
	}

	response := backend.DataResponse{}

	// for each query response, compile the data into response.Frames
	for _, singleResponse := range responseData {
		frame := singleResponse.ToFrame(qm.FormatOption)

		if config.UseLiveUpdate && qm.Live {
			channelFrame, err := createLiveChannel(singleResponse.PVname, frame, config.UID)
			if err != nil {
				log.DefaultLogger.Warn("Error applying live channel:", err)
			} else {
				frame.SetMeta(channelFrame)
			}
		}

		// add the frames to the response
		response.Frames = append(response.Frames, frame)
	}

	response.Error = responseErr

	return response
}

func applyAlias(sD []*models.SingleData, qm models.ArchiverQueryModel) ([]*models.SingleData, error) {
	// Alias is not set. Return data as is is.
	if qm.Alias == "" {
		return sD, nil
	}

	var rep *regexp.Regexp
	if qm.AliasPattern != "" {
		var err error
		rep, err = regexp.Compile(qm.AliasPattern)
		if err != nil {
			return sD, err
		}
	}

	for _, d := range sD {
		d.ApplyAlias(qm.Alias, rep)
	}

	return sD, nil
}

func dataExtrapol(singleResponse *models.SingleData, qm models.ArchiverQueryModel) *models.SingleData {
	if qm.Interval >= 1 || qm.Operator == "last" || qm.DisableExtrapol || qm.BackendQuery {
		return singleResponse
	}

	newResponse := singleResponse.Extrapolation(qm.TimeRange.To)

	return newResponse
}

func makeTargetPVList(client Client, target string, regex bool, maxNum int) []string {
	// PV name isolation for syntax like "(PV:NAME:1|PV:NAME:2|...)" is always required even if regex is enabled.
	// That's because AA sever doesn't support full regular expression.
	isolatedPvList := isolateBasicQuery(target)

	var targetPvList []string
	if regex {
		// If the user is using a regex to specify the PVs, parse and resolve the regex expression first
		// assemble the list of PVs to be queried for
		var regexPvList []string
		for _, v := range isolatedPvList {
			pvs, _ := client.FetchRegexTargetPVs(v, maxNum)
			regexPvList = append(regexPvList, pvs...)
		}
		targetPvList = regexPvList
	} else {
		targetPvList = isolatedPvList
	}

	// Each name in list should be unique
	var uniqPVList []string
	m := map[string]bool{}
	for _, pvname := range targetPvList {
		if !m[pvname] {
			m[pvname] = true
			uniqPVList = append(uniqPVList, pvname)
		}

	}

	return uniqPVList
}

func createLiveChannel(pvname string, _ *data.Frame, uuid string) (*data.FrameMeta, error) {
	//pvname := frame.Fields[1].Config.DisplayName
	valid := aalive.IsPVnameValid(pvname)

	if !valid {
		errMsg := fmt.Sprintf("createLiveChannel: invalid pvname %s", pvname)
		return nil, errors.New(errMsg)
	}

	var framemeta *data.FrameMeta
	if valid {
		path := aalive.ConvPV2URL(pvname)
		channel := live.Channel{
			Scope:     live.ScopeDatasource,
			Namespace: uuid,
			Path:      path,
		}
		framemeta = &data.FrameMeta{Channel: channel.String()}
	}

	return framemeta, nil
}

package main

import (
    "encoding/json"
    "net/http"
    "net/url"
    "time"
    "strings"
    "strconv"
    "io/ioutil"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

func IsBackendQuery(pluginctx backend.PluginContext) bool {
    // Return true if this query was created by the backend as opposed to visualization query for the frontend
    if pluginctx.User != nil {
        return true
    }
    return false
}

type ArchiverQueryModel struct {
    // It's not apparent to me where these two originate from but they do appear to be necessary
    Format string `json:"format"`
    Constant json.Number `json:"constant"` // I don't know what this is for yet
    QueryText string `json:"queryText"` // deprecated

    // Parameters added in AAQuery's extension of DataQuery
    Target string `json:"target"` //This will be the PV as entered by the user, or regex searching for PVs 
    Alias string `json:"alias"` // What to refer to the data as in the table - I think this only for the frontend rn
    AliasPattern string `json:"aliasPattern"` // use for collecting a large number of returned values 
    Operator string `json:"operator"` // ?
    Regex bool `json:"regex"` // configured by the user's setting of the "Regex" field in the panel
    Functions []FunctionDescriptorQueryModel `json:"functions"` // collection of functions to be applied to the data by the archiver

    // Only appears for visualization queries
    IntervalMs *int `json:"intervalMs,omitempty"`
    /*
        Functions is a few layers deep:
        Functions []FunctionDescriptorQueryModel contains: 
            Def FuncDefQueryModel contains:
                Params []FuncDefParamQueryModel
    */

    // Parameters from DataQuery
    RefId string `json:"refId"`
    Hide *bool `json:"hide"`
    Key *string `json:"string"`
    QueryType *string `json:"queryType"`
    DataTopic *string `json:"dataTopic"` //??
    Datasource *string `json:"datasource"` // comes back empty -- investigate further 
}

type FunctionDescriptorQueryModel struct {
    // Matched to FunctionDescriptor in types.ts
    Params []string `json:"params"`
    Def FuncDefQueryModel `json:"def"`
}

type FuncDefQueryModel struct {
    // Matched to FuncDef in types.ts
    DefaultParams *json.RawMessage `json:"defaultParams,omitempty"`
    ShortName *json.RawMessage `json:"shortName,omitempty"`
    Version *json.RawMessage `json:"version,omitempty"`
    Category string `json:"category"`
    Description *string `json:"description,omitempty"`
    Fake *bool `"json:"fake,omitempty"`
    Name string `json:"name"`
    Params []FuncDefParamQueryModel `json:"params"`
}

type FuncDefParamQueryModel struct {
    Name string `json"name"`
    Options *[]string `json:"options"`
    Type string `json:"type"`
}

func BuildQueryUrl(target string, query backend.DataQuery, pluginctx backend.PluginContext, qm ArchiverQueryModel) string {
    // Build the URL to query the archiver built from Grafana's configuration
    // Set some constants

    const TIME_FORMAT = "2006-01-02T15:04:05.000-07:00"
    const JSON_DATA_URL = "data/getData.qw"

    // Unpack the configured URL for the datasource and use that as the base for assembling the query URL
    u, err := url.Parse(pluginctx.DataSourceInstanceSettings.URL)
    if err != nil {
        log.DefaultLogger.Warn("err", "err", err)
    }

    // apply an operator to the PV string if one (not "raw" or "last") is provided
    opQuery, opErr := CreateOperatorQuery(qm)
    if opErr != nil {
        log.DefaultLogger.Warn("Operator has not been properly created")
    }

    // log.DefaultLogger.Debug("pluginctx","pluginctx", pluginctx)
    // log.DefaultLogger.Debug("query","query", query)

    var targetPv string
    if len(opQuery) > 0 {
        var opBuilder strings.Builder
        opBuilder.WriteString(opQuery)
        opBuilder.WriteString("(")
        opBuilder.WriteString(target)
        opBuilder.WriteString(")")
        targetPv = opBuilder.String()

    } else {
        targetPv = target
    }

    // amend the incomplete path
    var pathBuilder strings.Builder
    pathBuilder.WriteString(u.Path)
    pathBuilder.WriteString("/")
    pathBuilder.WriteString(JSON_DATA_URL)
    u.Path = pathBuilder.String()

    // assemble the query of the URL and attach it to u
    query_vals :=  make(url.Values)
    query_vals["pv"] = []string{targetPv}
    query_vals["from"] = []string{query.TimeRange.From.Format(TIME_FORMAT)}
    query_vals["to"] = []string{query.TimeRange.To.Format(TIME_FORMAT)}
    query_vals["donotchunk"] = []string{""}
    u.RawQuery = query_vals.Encode()

    // Display the result
    return u.String()
}

type SingleData struct {
    Name string
    Times []time.Time
    Values []float64
}

type ArchiverResponseModel struct {
    // Structure for unpacking the JSON response from the Archiver
    Meta struct{
        Name string `json:"name"`
        Waveform bool `json:"waveform"`
        EGU string `json:"EGU"`
        PREC json.Number `json:"PREC"`
    } `json:"meta"`
    Data []struct{
        Millis *json.Number`json:"millis,omitempty"`
        Nanos *json.Number`json:"nanos,omitempty"`
        Secs *json.Number`json:"secs,omitempty"`
        Val json.Number `json:"val"`
    } `json:"data"`
}

func ArchiverSingleQuery(queryUrl string) ([]byte, error){
    // Take the unformatted response from the http GET request and turn it into rows of timeseries data
    var jsonAsBytes []byte

    // Make the GET request
    httpResponse, getErr := http.Get(queryUrl)
    if getErr != nil {
        log.DefaultLogger.Warn("Get request has failed", "Error", getErr)
        return jsonAsBytes, getErr
    }

    // Convert get request response to variable and close the file
    jsonAsBytes, ioErr := ioutil.ReadAll(httpResponse.Body)
    httpResponse.Body.Close()
    if ioErr != nil {
        log.DefaultLogger.Warn("Parsing of incoming data has failed", "Error", ioErr)
        return jsonAsBytes, ioErr
    }

    return jsonAsBytes, nil
}

func ArchiverSingleQueryParser(jsonAsBytes []byte) (SingleData, error){
    // Convert received data to JSON
    var sD SingleData
    var data []ArchiverResponseModel
    jsonErr := json.Unmarshal(jsonAsBytes, &data)
    if jsonErr != nil {
        log.DefaultLogger.Warn("Conversion of incoming data to JSON has failed", "Error", jsonErr)
        return sD, jsonErr
    }

    // Obtain PV name 
    sD.Name = data[0].Meta.Name

    // Build output data block
    dataSize := len(data[0].Data)

    // initialize the slices with their final size so append operations are not necessary
    sD.Times = make([]time.Time, dataSize, dataSize)
    sD.Values = make([]float64, dataSize, dataSize)

    for idx, dataPt := range data[0].Data {

        millisCache, millisErr := dataPt.Millis.Int64()
        if millisErr != nil {
            log.DefaultLogger.Warn("Conversion of millis to int64 has failed", "Error", millisErr)
        }
        // use convert to nanoseconds
        sD.Times[idx] = time.Unix(0, 1e6 * millisCache)
        valCache, valErr := dataPt.Val.Float64()
        if valErr != nil {
            log.DefaultLogger.Warn("Conversion of val to float64 has failed", "Error", valErr)
        }
        sD.Values[idx] = valCache
    }
    return sD, nil
}

func BuildRegexUrl(regex string, pluginctx backend.PluginContext) string {
    // Construct the request URL for the regex search of PVs and return it as a string
    const REGEX_URL = "bpl/getMatchingPVs"
    const REGEX_MAXIMUM_MATCHES = 1000

    // Unpack the configured URL for the datasource and use that as the base for assembling the query URL
    u, err := url.Parse(pluginctx.DataSourceInstanceSettings.URL)
        if err != nil {
        log.DefaultLogger.Warn("err", "err", err)
    }

    // amend the incomplete path
    var pathBuilder strings.Builder
    pathBuilder.WriteString(u.Path)
    pathBuilder.WriteString("/")
    pathBuilder.WriteString(REGEX_URL)
    u.Path = pathBuilder.String()

    // assemble the query of the URL and attach it to u
    query_vals :=  make(url.Values)
    query_vals["regex"] = []string{regex}
    query_vals["limit"] = []string{strconv.Itoa(REGEX_MAXIMUM_MATCHES)}
    u.RawQuery = query_vals.Encode()

    // log.DefaultLogger.Debug("u.String", "value", u.String())
    return u.String()
}

func ArchiverRegexQuery(queryUrl string) ([]byte, error){
    // Make the GET request  for the JSON list of matching PVs, parse it, and return a list of strings
    var jsonAsBytes []byte

    // Make the GET request
    httpResponse, getErr := http.Get(queryUrl)
    if getErr != nil {
        log.DefaultLogger.Warn("Get request has failed", "Error", getErr)
        return jsonAsBytes, getErr
    }

    // Convert get request response to variable and close the file
    jsonAsBytes, ioErr := ioutil.ReadAll(httpResponse.Body)
    httpResponse.Body.Close()
    if ioErr != nil {
        log.DefaultLogger.Warn("Parsing of incoming data has failed", "Error", ioErr)
        return jsonAsBytes, ioErr
    }
    return jsonAsBytes, nil
}

func ArchiverRegexQueryParser(jsonAsBytes []byte) ([]string, error){
    // Convert received data to JSON
    var pvList []string
    jsonErr := json.Unmarshal(jsonAsBytes, &pvList)
    if jsonErr != nil {
        log.DefaultLogger.Warn("Conversion of incoming data to JSON has failed", "Error", jsonErr)
        return pvList, jsonErr
    }

    return pvList, nil
}

func ExecuteSingleQuery(target string, query backend.DataQuery, pluginctx backend.PluginContext, qm ArchiverQueryModel) (SingleData, error) {
    // wrap together the individual operations build a query, execute the query, and compile the data into a singleData structure
    // target: This is the PV to be queried for. As the "query" argument may be a regular expression, the specific PV desired must be specified
    queryUrl := BuildQueryUrl(target, query, pluginctx, qm)
    queryResponse, _ := ArchiverSingleQuery(queryUrl)
    parsedResponse, _ := ArchiverSingleQueryParser(queryResponse)
    return parsedResponse, nil
}

func IsolateBasicQuery(unparsed string) []string {
    // Non-regex queries can request multiple PVs using this syntax: (PV:NAME:1|PV:NAME:2|...)
    // This function takes queries in this format and breaks them up into a list of individual PVs
    unparsed_clean := strings.TrimSpace(unparsed)

    phrases := LocateOuterParen(unparsed_clean)
    // Identify parenthesis-bound sections
    multiPhrases := phrases.Phrases
    // Locate parenthesis-bound sections
    phraseIdxs := phrases.Idxs

    // If there are no sub-phrases in this string, return immediately to prevent further recursion
    if len(multiPhrases) == 0 {
        return []string{unparsed_clean}
    }

    // A list of all the possible phrases
    phraseParts := make([][]string, 0, len(multiPhrases))

    for idx, _ := range multiPhrases {
        // Strip leading and ending parenthesis
        multiPhrases[idx] = multiPhrases[idx][1:len(multiPhrases[idx])-1]
        // Break parsed phrases on "|"
        phraseParts = append(phraseParts, SplitLowestLevelOnly(multiPhrases[idx]))
    }

    // list of all the configurations for the in-order phrases to be inserted
    phraseCase := PermuteQuery(phraseParts)

    result := make([]string, 0, len(phraseCase))

    // Build results by substituting all phrase combinations in place for 1st-level substitutions 
    for _, phrase := range phraseCase {
        createdString := SelectiveInsert(unparsed_clean, phraseIdxs, phrase)
        result = append(result, createdString)
    }

    // For any phrase that has sub-phrases in need of parsing, call this function again on the sub-phrase and append the results to the end of the current output.
    for pos, chunk := range result {
        parseAttempt := IsolateBasicQuery(chunk)
        if len(parseAttempt) > 1 {
            result = append(result[:pos], result[pos+1:]...) // pop partially parsed entry
            result = append(result, parseAttempt...) // add new entires at the end of the list. 
        }
    }

    return result
}

func SplitLowestLevelOnly(inputData string) []string {
    output := make([]string,0,5)
    nestCounter := 0
    stashInitPos := 0
    for pos, char := range inputData {
        switch {
            case char == '(':
                nestCounter++
            case char == ')':
                nestCounter--
        }
        if char == '|' && nestCounter == 0 {
            output = append(output, inputData[stashInitPos:pos])
            stashInitPos = pos+1
        }
    }
    output = append(output, inputData[stashInitPos:])
    return output 
}

type ParenLoc struct {
    // Use to identify unenenclosed parenthesis, and their content
    Phrases []string
    Idxs [][]int
}

func LocateOuterParen(inputData string) ParenLoc {
    // read through a string to identify all paired sets of parentheses that are not contained within another set of parenthesis. 
    // When found, report the indexes of all instantces as well as the content contained within the parenthesis.
    // This function ignores internal parenthesis. 
    var output ParenLoc;

    nestCounter := 0
    var stashInitPos int
    for pos, char := range inputData {
        if char == '(' {
            if nestCounter == 0 {
                stashInitPos = pos
            }
            nestCounter++
        }
        if char == ')' {
            if nestCounter == 1 {
                // A completed 1st-level parenthesis set has been completed 
                output.Phrases = append(output.Phrases, inputData[stashInitPos:pos+1])
                output.Idxs = append(output.Idxs, []int{stashInitPos, pos+1})
            }
            nestCounter--
        }
    }
    return output
}

func PermuteQuery( inputData [][]string) [][]string {
    /*
        Generate all ordered permutations of the input strings to make the following operation occur: 

        input:
            {
                {"a", "b"}
                {"c", "d"}
            }

        output:
            {
                {"a", "c"}
                {"a", "d"}
                {"b", "c"}
                {"b", "d"}
            }
    */
    return permuteQueryRecurse(inputData,[]string{}) 
}

func permuteQueryRecurse( inputData [][]string, trace []string) [][]string {
    /* 
        recursive method for visitng all the permutations. 
    */

    // If you've assigned a value for each phrase, just return the full sequence of phrases 
    if len(trace) == len(inputData) {
        output := make([][]string,0,1)
        output = append(output, trace)
        return output
    }

    /*
        When values aren't assigned to all phrases, begin by assigning the first unassigned value 
        to each of the possible values and recursing on the result of that
    */
    targetIdx := len(trace)
    output := make([][]string, 0, len(inputData[targetIdx]))
    for _, value := range inputData[targetIdx] {
        response := permuteQueryRecurse(inputData, append(trace, value))
        for _, oneOutput := range response {
            output = append(output, oneOutput)
        }
    }
    return output
}

func SelectiveInsert( input string, idxs [][]int, inserts []string) string {
    /*
        Selectively replace portions of the input string

        idxs indicates the indices which will be removed

        inserts provides the new string that will be put into the input string at the place designated by idxs

        idxs and inserts should have the same length and are matched with each other element-wise to support any number of substitutions
    */
    var builder strings.Builder

    prevIdx := 0

    // return a blank string if the arguments are bad
    if len(idxs) != len(inserts) {
        return ""
    }

    // At each place marked by idx, insert the corresponding new string
    for idx, val := range idxs {
        firstVal := val[0]
        builder.WriteString(input[prevIdx:firstVal])
        builder.WriteString(inserts[idx])
        prevIdx = val[1]
    }

    // Handle any trailing string 
    if prevIdx < len(input) {
        lastVal := len(input)
        builder.WriteString(input[prevIdx:lastVal])
    }

    return builder.String()
}

func FrameBuilder(singleResponse SingleData) *data.Frame {
        // create data frame response
        frame := data.NewFrame("response")

        //add the time dimension
        frame.Fields = append(frame.Fields,
            data.NewField("time", nil, singleResponse.Times),
        )

        // add values 
        frame.Fields = append(frame.Fields,
            data.NewField(singleResponse.Name, nil, singleResponse.Values),
        )

        return frame
}

func DataExtrapol(singleResponse SingleData, qm ArchiverQueryModel, query backend.DataQuery) SingleData {
    disableExtrapol, err := qm.DisableExtrapol()
    if err != nil {
        disableExtrapol = false
    }

    if (qm.Operator != "raw") || disableExtrapol {
        return singleResponse
    }

    newResponse := singleResponse
    newValue := singleResponse.Values[len(singleResponse.Values)-1]
    newTime := query.TimeRange.To

    newResponse.Values = append(newResponse.Values, newValue)
    newResponse.Times = append(newResponse.Times, newTime)

    return newResponse
}

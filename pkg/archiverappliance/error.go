package archiverappliance

import "errors"

var (
	errEmptyResponse         = errors.New("response is empty")
	errIllegalPayloadType    = errors.New("response from Archiver Appliance might be illegal payload type")
	errFailedToParsePBFormat = errors.New("failed to parse the PB format response")
)

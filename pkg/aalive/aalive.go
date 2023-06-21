package aalive

import (
	"regexp"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

const (
	PV_SPLIITTER = ":"
	URL_SPLITTER = "="
)

var pvreg = regexp.MustCompile(`^[a-zA-Z0-9\.\-_:].*$`)
var urlreg = regexp.MustCompile(`^[a-zA-Z0-9\.\-_=].*$`)

func ConvPV2URL(pvname string) string {
	return strings.Replace(pvname, PV_SPLIITTER, URL_SPLITTER, -1)
}

func ConvURL2PV(url string) string {
	return strings.Replace(url, URL_SPLITTER, PV_SPLIITTER, -1)
}

func IsPVnameValid(pvname string) bool {
	return pvreg.MatchString(pvname)
}

func IsPathValid(pvname string) bool {
	return urlreg.MatchString(pvname)
}

func SendErrorFrame(msg string, sender *backend.StreamSender) {
	frame := data.NewFrame("error")
	frame.Fields = append(frame.Fields, data.NewField("error", nil, []string{msg}))

	serr := sender.SendFrame(frame, data.IncludeAll)
	if serr != nil {
		log.DefaultLogger.Error("Failed to send error frame", "error", serr)
	}
}

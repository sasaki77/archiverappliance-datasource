package aalive

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"nhooyr.io/websocket"
)

type wsDataProxy struct {
	wsUrl         string
	msgRead       chan []byte
	sender        *backend.StreamSender
	pv            string
	ctx           context.Context
	Done          chan bool
	ReadingErrors chan error
}

func NewWsDataProxy(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender, pv string) (*wsDataProxy, error) {
	wsDataProxy := &wsDataProxy{
		msgRead:       make(chan []byte),
		sender:        sender,
		pv:            pv,
		ctx:           ctx,
		Done:          make(chan bool, 1),
		ReadingErrors: make(chan error),
	}

	url, err := wsDataProxy.encodeURL(pv)
	if err != nil {
		return nil, fmt.Errorf("encode URL Error: %s", err.Error())
	}
	wsDataProxy.wsUrl = url
	log.DefaultLogger.Info(url)

	return wsDataProxy, nil
}

func (wsdp *wsDataProxy) ReadMessage() {
	ctx := context.Background()

	c, _, _ := websocket.Dial(ctx, wsdp.wsUrl, nil)
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	log.DefaultLogger.Info("Ws Connect", "connected to", wsdp.wsUrl)

	b := []byte(`{ "type": "subscribe", "pvs": [ "ET_SASAKI:TEST" ] }`)
	err := c.Write(ctx, 1, b)
	if err != nil {
		log.DefaultLogger.Info("ERROR")
	}

	for {
		select {
		case <-wsdp.Done:
			return
		default:
			_, v, err := c.Read(ctx)
			log.DefaultLogger.Info(string(v))
			if err != nil {
				log.DefaultLogger.Info(err.Error())
				time.Sleep(3 * time.Second)
				wsdp.ReadingErrors <- fmt.Errorf("%s: %s", "Error reading the websocket", err.Error())
				return
			} else {
				wsdp.msgRead <- v
			}
		}
	}
}

type messageModel struct {
	Value   float64 `json:"value"`
	Seconds int64   `json:"seconds"`
	Nanos   int64   `json:"nanos"`
}

func (wsdp *wsDataProxy) ProxyMessage() {
	frame := data.NewFrame(wsdp.pv)

	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, make([]time.Time, 1)),
		data.NewField(wsdp.pv, nil, make([]float64, 1)),
	)

	m := messageModel{}

	for {
		message, ok := <-wsdp.msgRead
		// if channel was closed
		if !ok {
			return
		}

		json.Unmarshal(message, &m)

		t := time.Unix(m.Seconds, m.Nanos)
		frame.Fields[0].Set(0, t)
		frame.Fields[1].Set(0, m.Value)

		err := wsdp.sender.SendFrame(frame, data.IncludeAll)
		if err != nil {
			log.DefaultLogger.Error("Failed to send frame", "error", err)
		}
	}
}

// encodeURL is hard coded with some variables like scheme and x-api-key but will be definetly refactored after changes in the config editor
func (wsdp *wsDataProxy) encodeURL(req string) (string, error) {
	host := "ws://172.20.240.1:8080/pvws/pv"

	wsUrl, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("failed to parse host string from the Plugin's Config Editor: %s", err.Error())
	}

	wsUrl.Path = path.Join(wsUrl.Path)

	//return wsUrl.String(), nil
	return host, nil
}

func SendErrorFrame(msg string, sender *backend.StreamSender) {
	frame := data.NewFrame("error")
	frame.Fields = append(frame.Fields, data.NewField("error", nil, []string{msg}))

	serr := sender.SendFrame(frame, data.IncludeAll)
	if serr != nil {
		log.DefaultLogger.Error("Failed to send error frame", "error", serr)
	}
}

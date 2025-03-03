package aalive

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"nhooyr.io/websocket"
)

type wsDataProxy struct {
	wsUrl         string
	wsConn        *websocket.Conn
	msgRead       chan []byte
	sender        *backend.StreamSender
	pvname        string
	ReadingErrors chan error
	Done          chan bool
}

func NewWsDataProxy(ctx context.Context, sender *backend.StreamSender, pvname string, uri string) (*wsDataProxy, error) {
	wsDataProxy := &wsDataProxy{
		msgRead:       make(chan []byte),
		sender:        sender,
		pvname:        pvname,
		ReadingErrors: make(chan error),
		Done:          make(chan bool, 1),
	}

	url, err := wsDataProxy.encodeURL(uri)
	if err != nil {
		return nil, fmt.Errorf("encode URL Error: %s", err.Error())
	}
	wsDataProxy.wsUrl = url
	log.DefaultLogger.Debug(url)

	c, err := wsDataProxy.wsConnect(ctx)
	if err != nil {
		return nil, fmt.Errorf("connection Error: %s", err.Error())
	}
	wsDataProxy.wsConn = c

	return wsDataProxy, nil
}

func (wsdp *wsDataProxy) ReadMessage() {
	defer func() {
		wsdp.wsConn.Close(websocket.StatusNormalClosure, "")
		close(wsdp.msgRead)
		log.DefaultLogger.Debug("Read Message routine", "detail", "closing websocket connection and msgRead channel")
	}()

	ctx := context.Background()

	// Start Subscribe
	b := []byte(fmt.Sprintf(`{ "type": "subscribe", "pvs": [ "%s" ] }`, wsdp.pvname))
	err := wsdp.wsConn.Write(ctx, 1, b)
	if err != nil {
		wsdp.ReadingErrors <- fmt.Errorf("%s: %s", "Error writing the websocket", err.Error())
		return
	}

	go func() {
		for {
			_, v, err := wsdp.wsConn.Read(ctx)
			if err != nil {
				time.Sleep(3 * time.Second)
				wsdp.ReadingErrors <- fmt.Errorf("%s: %s", "Error reading the websocket", err.Error())
				return
			}
			wsdp.msgRead <- v
		}
	}()

	<-wsdp.Done
	log.DefaultLogger.Debug("ReadMessage closed")
}

type messageModel struct {
	Value   float64 `json:"value"`
	Seconds int64   `json:"seconds"`
	Nanos   int64   `json:"nanos"`
}

func (wsdp *wsDataProxy) ProxyMessage() {
	frame := data.NewFrame(wsdp.pvname)

	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, make([]time.Time, 1)),
		data.NewField(wsdp.pvname, nil, make([]float64, 1)),
	)

	m := messageModel{}

	for {
		message, ok := <-wsdp.msgRead
		// if channel was closed
		if !ok {
			log.DefaultLogger.Debug("ProxyMessage closed")
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

func (wsdp *wsDataProxy) encodeURL(uri string) (string, error) {
	host := uri

	return host, nil
}

func (wsdp *wsDataProxy) wsConnect(ctx context.Context) (*websocket.Conn, error) {
	log.DefaultLogger.Debug("Ws Connect", "connecting to", wsdp.wsUrl)

	c, _, err := websocket.Dial(ctx, wsdp.wsUrl, nil)
	if err != nil {
		return nil, err
	}
	log.DefaultLogger.Debug("Ws Connect", "connected to", wsdp.wsUrl)

	return c, nil
}

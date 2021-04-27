package dispatcher

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/demoware/ingester/payload_handler"
	"github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

type DispatcherOptions struct {
	// The maximum number of metrics that can be processed by all the handlers
	BatchLimit uint
	// The different components used in handling the payload data
	PayloadHandlers []payload_handler.PayloadHandler
	// The server uri to the demoware metrics producer
	MetricsUri string
	// The auth token to the server
	AuthToken string
}

type Dispatcher struct {
	client          *http.Client
	BatchLimit      uint
	PayloadHandlers []payload_handler.PayloadHandler
	MetricsUri      string
	AuthToken       string
}

func (o *DispatcherOptions) NewDispatcher() *Dispatcher {
	if o.BatchLimit == 0 {
		o.BatchLimit = 5
	}
	initMetrics()
	return &Dispatcher{
		client: &http.Client{
			Timeout: 1 * time.Second,
		},
		BatchLimit:      o.BatchLimit,
		PayloadHandlers: o.PayloadHandlers,
		MetricsUri:      o.MetricsUri,
		AuthToken:       o.AuthToken,
	}
}

func (d *Dispatcher) NewHandler(handler payload_handler.PayloadHandler) {
	d.PayloadHandlers = append(d.PayloadHandlers, handler)
}

func (d *Dispatcher) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if d.CanProcessNewPayload() {
				err := d.ProcessPayload()
				if err != nil {
					log.Errorf("error occured when processing payload: %e", err)
				}
			}
		}
	}
}

func (d *Dispatcher) ProcessPayload() error {
	req, _ := http.NewRequest("GET", d.MetricsUri, nil)
	req.SetBasicAuth(d.AuthToken, "")
	req.Header.Set("Accept", "application/json")
	resp, err := d.client.Do(req)

	if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
		increaseErrorCount()
		return errors.New("error occurred getting metrics from server")
	}

	if resp.StatusCode == http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		var data []payload_handler.Payload
		err = json.Unmarshal(body, &data)
		if err != nil {
			increaseErrorCount()
			return errors.New("payload data in wrong format")
		}

		updateProcessPayloadMetrics(len(data))

		for _, dt := range data {
			handler := d.GetHandler(dt.Type)
			if handler != nil {
				handler.InsertPayload(dt)
			}
		}
	}
	return nil
}

func (d *Dispatcher) CanProcessNewPayload() bool {
	for _, h := range d.PayloadHandlers {
		if h.CountUnProcessedPayload() > d.BatchLimit {
			return false
		}
	}
	return true
}

func (d *Dispatcher) GetHandler(payloadType string) payload_handler.PayloadHandler {
	for _, h := range d.PayloadHandlers {
		if h.GetHandlerType() == payloadType {
			return h
		}
	}
	return nil
}

func initMetrics() {
	c := metrics.NewCounter()
	_ = metrics.Register("dispatcher_payload_processed_count", c)
	e := metrics.NewCounter()
	_ = metrics.Register("dispatcher_errors", e)
}

func increaseErrorCount() {
	counter := metrics.DefaultRegistry.Get("dispatcher_errors")
	counter.(metrics.Counter).Inc(1)
}

func updateProcessPayloadMetrics(count int) {
	counter := metrics.DefaultRegistry.Get("dispatcher_payload_processed_count")
	counter.(metrics.Counter).Inc(int64(count))
}

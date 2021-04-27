package load_average_handler

import (
	"context"
	"errors"
	"github.com/demoware/ingester/payload_handler"
	"github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
)

const (
	handlerType = "load_avg"
)

type LoadAverageHandler struct {
	minimumLoad        float64
	maximumLoad        float64
	unprocessedPayload uint
	payloads           chan payload_handler.Payload
}

func NewLoadAverageHandler() *LoadAverageHandler {
	initMetrics()
	return &LoadAverageHandler{
		payloads: make(chan payload_handler.Payload, 10),
	}
}

func initMetrics() {
	c := metrics.NewCounter()
	_ = metrics.Register(handlerType+"_payload_processed_count", c)
	e := metrics.NewCounter()
	_ = metrics.Register(handlerType+"_errors", e)
	mig := metrics.NewGaugeFloat64()
	_ = metrics.Register(handlerType+"_minimum_load", mig)
	mag := metrics.NewGaugeFloat64()
	_ = metrics.Register(handlerType+"_maximum_load", mag)
}

func (l *LoadAverageHandler) InsertPayload(payload payload_handler.Payload) {
	l.payloads <- payload
}

func (l *LoadAverageHandler) GetHandlerType() string {
	return handlerType
}

func (l *LoadAverageHandler) CountUnProcessedPayload() uint {
	return l.unprocessedPayload
}

func (l *LoadAverageHandler) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(l.payloads)
			return
		default:
			err := l.ProcessLoadAveragePayload()
			if err != nil {
				log.Errorf("error occured when processing payload: %e", err)
			}
		}
	}
}

func (l *LoadAverageHandler) ProcessLoadAveragePayload() error {
	val, ok := <-l.payloads
	if !ok {
		increaseErrorCount()
		return errors.New("wrong load average payload")
	}
	payload, ok := val.Payload.(map[string]interface{})
	if !ok || payload["value"] == nil {
		increaseErrorCount()
		return errors.New("wrong load average payload")
	}
	data, ok := payload["value"].(interface{}).(float64)
	if !ok {
		increaseErrorCount()
		return errors.New("wrong load average payload")
	}
	l.unprocessedPayload++
	if data < l.minimumLoad || l.minimumLoad == 0 {
		l.minimumLoad = data
	}
	if data > l.maximumLoad {
		l.maximumLoad = data
	}
	l.unprocessedPayload--
	l.updateMetrics()
	return nil
}

func (l *LoadAverageHandler) updateMetrics() {
	counter := metrics.DefaultRegistry.Get(handlerType + "_payload_processed_count")
	counter.(metrics.Counter).Inc(1)
	minGauge := metrics.DefaultRegistry.Get(handlerType + "_minimum_load")
	minGauge.(metrics.GaugeFloat64).Update(l.minimumLoad)
	maxGauge := metrics.DefaultRegistry.Get(handlerType + "_maximum_load")
	maxGauge.(metrics.GaugeFloat64).Update(l.maximumLoad)
}

func increaseErrorCount() {
	counter := metrics.DefaultRegistry.Get(handlerType + "_errors")
	counter.(metrics.Counter).Inc(1)
}

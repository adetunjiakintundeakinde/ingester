package last_kernel_update_handler

import (
	"context"
	"errors"
	"github.com/demoware/ingester/payload_handler"
	"github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	handlerType = "last_kernel_upgrade"
)

type LastKernelUpgradeHandler struct {
	latestUpgrade      time.Time
	unprocessedPayload uint
	payloads           chan payload_handler.Payload
}

func NewLastKernelUpgradeHandler() *LastKernelUpgradeHandler {
	initMetrics()
	return &LastKernelUpgradeHandler{
		payloads: make(chan payload_handler.Payload, 10),
	}
}

func initMetrics() {
	c := metrics.NewCounter()
	_ = metrics.Register(handlerType+"_payload_processed_count", c)
	e := metrics.NewCounter()
	_ = metrics.Register(handlerType+"_errors", e)
	g := metrics.NewGauge()
	_ = metrics.Register(handlerType+"_latest_time", g)
}

func (l *LastKernelUpgradeHandler) InsertPayload(payload payload_handler.Payload) {
	l.payloads <- payload
}

func (l *LastKernelUpgradeHandler) GetHandlerType() string {
	return handlerType
}

func (l *LastKernelUpgradeHandler) CountUnProcessedPayload() uint {
	return l.unprocessedPayload
}

func (l *LastKernelUpgradeHandler) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(l.payloads)
			return
		default:
			err := l.ProcessLastKernelUpdatePayload()
			if err != nil {
				log.Errorf("error occured when processing payload: %e", err)
			}
		}
	}
}

func (l *LastKernelUpgradeHandler) ProcessLastKernelUpdatePayload() error {
	val, ok := <-l.payloads
	if !ok {
		increaseErrorCount()
	}

	payload, ok := val.Payload.(map[string]interface{})
	if !ok || payload["value"] == nil {
		increaseErrorCount()
		return errors.New("wrong last kernel update payload")
	}

	data, ok := payload["value"].(interface{}).(string)
	if !ok {
		increaseErrorCount()
		return errors.New("wrong last kernel update payload")
	}
	newTime, err := time.Parse(time.RFC3339Nano, data)
	if err != nil {
		increaseErrorCount()
		return errors.New("wrong last kernel update payload")
	}
	l.unprocessedPayload++
	if l.latestUpgrade.IsZero() {
		l.latestUpgrade = newTime
	} else if l.latestUpgrade.Before(newTime) {
		l.latestUpgrade = newTime
	}
	l.unprocessedPayload--
	l.updateMetrics()
	return nil
}

func (l *LastKernelUpgradeHandler) updateMetrics() {
	counter := metrics.DefaultRegistry.Get(handlerType + "_payload_processed_count")
	counter.(metrics.Counter).Inc(1)
	gauge := metrics.DefaultRegistry.Get(handlerType + "_latest_time")
	gauge.(metrics.Gauge).Update(l.latestUpgrade.Unix())
}

func increaseErrorCount() {
	counter := metrics.DefaultRegistry.Get(handlerType + "_errors")
	counter.(metrics.Counter).Inc(1)
}

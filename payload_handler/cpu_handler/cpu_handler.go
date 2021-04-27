package cpu_handler

import (
	"context"
	"errors"
	"github.com/demoware/ingester/payload_handler"
	"github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
)

const (
	handlerType = "cpu_usage"
)

type CpuHandler struct {
	averageCPU         float64
	counter            uint64
	unprocessedPayload uint
	payloads           chan payload_handler.Payload
}

func NewCpuHandler() *CpuHandler {
	initMetrics()
	return &CpuHandler{
		payloads: make(chan payload_handler.Payload, 10),
	}
}

func initMetrics() {
	c := metrics.NewCounter()
	_ = metrics.Register(handlerType+"_payload_processed_count", c)
	e := metrics.NewCounter()
	_ = metrics.Register(handlerType+"_errors", e)
	g := metrics.NewGaugeFloat64()
	_ = metrics.Register(handlerType+"_average_cpu", g)
}

func (c *CpuHandler) InsertPayload(payload payload_handler.Payload) {
	c.payloads <- payload
}

func (c *CpuHandler) GetHandlerType() string {
	return handlerType
}

func (c *CpuHandler) CountUnProcessedPayload() uint {
	return c.unprocessedPayload
}

func (c *CpuHandler) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// when
			close(c.payloads)
			return
		default:
			err := c.ProcessCpuPayload()
			if err != nil {
				log.Errorf("error occured when processing payload: %e", err)
			}
		}
	}
}

func (c *CpuHandler) ProcessCpuPayload() error {
	val, ok := <-c.payloads
	if !ok {
		increaseErrorCount()
		return errors.New("could not retrieve payload data from channel")
	}
	payload, ok := val.Payload.(map[string]interface{})
	if !ok || payload["value"] == nil {
		increaseErrorCount()
		return errors.New("wrong cpu payload data")
	}
	data, ok := payload["value"].([]interface{})
	if !ok {
		increaseErrorCount()
		return errors.New("wrong cpu payload data")
	}
	c.unprocessedPayload++
	for _, dt := range data {
		f, ok := dt.(float64)
		if !ok {
			/*
				Break the loop of part of the data is incorrect
			*/
			log.Errorf("cpu data %s has wrong data type", dt)
			increaseErrorCount()
			return errors.New("wrong cpu payload data")
		}
		c.counter++
		previousCount := c.counter - 1
		c.averageCPU = ((float64(previousCount) * c.averageCPU) + f) / float64(c.counter)
	}
	c.unprocessedPayload--
	c.updateMetrics()
	return nil
}

func (c *CpuHandler) updateMetrics() {
	counter := metrics.DefaultRegistry.Get(handlerType + "_payload_processed_count")
	counter.(metrics.Counter).Inc(1)
	gauge := metrics.DefaultRegistry.Get(handlerType + "_average_cpu")
	gauge.(metrics.GaugeFloat64).Update(c.averageCPU)
}

func increaseErrorCount() {
	counter := metrics.DefaultRegistry.Get(handlerType + "_errors")
	counter.(metrics.Counter).Inc(1)
}

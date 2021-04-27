package cpu_handler

import (
	"github.com/demoware/ingester/payload_handler"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCpuHandler_GetHandlerType(t *testing.T) {
	cpuHandler := NewCpuHandler()
	require.Equal(t, cpuHandler.GetHandlerType(), "cpu_usage")
}

func TestCpuHandler_CountUnProcessedPayload(t *testing.T) {
	cpuHandler := CpuHandler{
		unprocessedPayload: 5,
	}
	require.Equal(t, cpuHandler.CountUnProcessedPayload(), uint(5))
}

func TestCpuHandler_ProcessCpuPayload(t *testing.T) {
	cpuHandler := NewCpuHandler()
	var payload = payload_handler.Payload{
		Type:    "cpu_handler",
		Payload: map[string]interface{}{"value": []interface{}{2.5, 1.5, 5.0}},
	}
	cpuHandler.InsertPayload(payload)
	require.NoError(t, cpuHandler.ProcessCpuPayload())
	require.Equal(t, cpuHandler.averageCPU, float64(3))
	require.Equal(t, cpuHandler.counter, uint64(3))
}

func TestCpuHandler_ProcessCpuPayload_Rolling_Average(t *testing.T) {
	cpuHandler := NewCpuHandler()
	var payload = payload_handler.Payload{
		Type:    "cpu_handler",
		Payload: map[string]interface{}{"value": []interface{}{2.5, 1.5, 5.0}},
	}
	cpuHandler.InsertPayload(payload)
	require.NoError(t, cpuHandler.ProcessCpuPayload())
	require.Equal(t, cpuHandler.averageCPU, float64(3))
	require.Equal(t, cpuHandler.counter, uint64(3))

	var newPayload = payload_handler.Payload{
		Type:    "cpu_handler",
		Payload: map[string]interface{}{"value": []interface{}{6.0, 5.0}},
	}
	cpuHandler.InsertPayload(newPayload)
	require.NoError(t, cpuHandler.ProcessCpuPayload())
	require.Equal(t, cpuHandler.averageCPU, float64(4))
	require.Equal(t, cpuHandler.counter, uint64(5))
}

func TestCpuHandler_ProcessCpuPayload_IncorrectPayload(t *testing.T) {
	cpuHandler := NewCpuHandler()
	var payload = payload_handler.Payload{
		Type:    "cpu_handler",
		Payload: map[string]interface{}{"value": []interface{}{2.5, "test", 5.0}},
	}
	cpuHandler.InsertPayload(payload)
	require.Error(t, cpuHandler.ProcessCpuPayload())
}

func TestCpuHandler_ProcessCpuPayload_IncorrectPayload_NoValue(t *testing.T) {
	cpuHandler := NewCpuHandler()
	var payload = payload_handler.Payload{
		Type:    "cpu_handler",
		Payload: map[string]interface{}{"values": []interface{}{2.5, "test", 5.0}},
	}
	cpuHandler.InsertPayload(payload)
	require.Error(t, cpuHandler.ProcessCpuPayload())
}

package load_average_handler

import (
	"github.com/demoware/ingester/payload_handler"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLoadAverageHandler_GetHandlerType(t *testing.T) {
	lastKernelUpgradeHandler := NewLoadAverageHandler()
	require.Equal(t, lastKernelUpgradeHandler.GetHandlerType(), "load_avg")
}

func TestLoadAverageHandler_CountUnProcessedPayload(t *testing.T) {
	loadAverageHandler := LoadAverageHandler{
		unprocessedPayload: 5,
	}
	require.Equal(t, loadAverageHandler.CountUnProcessedPayload(), uint(5))
}

func TestLastKernelUpgradeHandler_ProcessLastKernelUpgradePayload_InvalidData(t *testing.T) {
	lastKernelUpgradeHandler := NewLoadAverageHandler()
	var payload = payload_handler.Payload{
		Type:    "load_avg",
		Payload: map[string]interface{}{"value": "rubbish"},
	}
	lastKernelUpgradeHandler.InsertPayload(payload)
	require.Error(t, lastKernelUpgradeHandler.ProcessLoadAveragePayload())
}

func TestLastKernelUpgradeHandler_ProcessLastKernelUpgradePayload(t *testing.T) {
	lastKernelUpgradeHandler := NewLoadAverageHandler()
	var payload = payload_handler.Payload{
		Type:    "load_avg",
		Payload: map[string]interface{}{"value": 5.0},
	}
	lastKernelUpgradeHandler.InsertPayload(payload)
	require.NoError(t, lastKernelUpgradeHandler.ProcessLoadAveragePayload())
	require.Equal(
		t,
		lastKernelUpgradeHandler.maximumLoad,
		float64(5.0),
	)
	require.Equal(
		t,
		lastKernelUpgradeHandler.minimumLoad,
		float64(5.0),
	)
}

func TestLastKernelUpgradeHandler_ProcessLastKernelUpgradePayload_RollingChange(t *testing.T) {
	lastKernelUpgradeHandler := NewLoadAverageHandler()
	var payload = payload_handler.Payload{
		Type:    "load_avg",
		Payload: map[string]interface{}{"value": 5.0},
	}
	lastKernelUpgradeHandler.InsertPayload(payload)
	require.NoError(t, lastKernelUpgradeHandler.ProcessLoadAveragePayload())
	require.Equal(
		t,
		lastKernelUpgradeHandler.maximumLoad,
		float64(5.0),
	)
	require.Equal(
		t,
		lastKernelUpgradeHandler.minimumLoad,
		float64(5.0),
	)

	var newPayload = payload_handler.Payload{
		Type:    "load_avg",
		Payload: map[string]interface{}{"value": 2.0},
	}
	lastKernelUpgradeHandler.InsertPayload(newPayload)
	require.NoError(t, lastKernelUpgradeHandler.ProcessLoadAveragePayload())
	require.Equal(
		t,
		lastKernelUpgradeHandler.maximumLoad,
		float64(5.0),
	)
	require.Equal(
		t,
		lastKernelUpgradeHandler.minimumLoad,
		float64(2.0),
	)

	var anotherPayload = payload_handler.Payload{
		Type:    "load_avg",
		Payload: map[string]interface{}{"value": 7.0},
	}
	lastKernelUpgradeHandler.InsertPayload(anotherPayload)
	require.NoError(t, lastKernelUpgradeHandler.ProcessLoadAveragePayload())
	require.Equal(
		t,
		lastKernelUpgradeHandler.maximumLoad,
		float64(7.0),
	)
	require.Equal(
		t,
		lastKernelUpgradeHandler.minimumLoad,
		float64(2.0),
	)
}

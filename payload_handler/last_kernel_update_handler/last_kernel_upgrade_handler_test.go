package last_kernel_update_handler

import (
	"github.com/demoware/ingester/payload_handler"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestLastKernelUpgradeHandler_GetHandlerType(t *testing.T) {
	lastKernelUpgradeHandler := NewLastKernelUpgradeHandler()
	require.Equal(t, lastKernelUpgradeHandler.GetHandlerType(), "last_kernel_upgrade")
}

func TestLastKernelUpgradeHandler_CountUnProcessedPayload(t *testing.T) {
	lastKernelUpgradeHandler := LastKernelUpgradeHandler{
		unprocessedPayload: 5,
	}
	require.Equal(t, lastKernelUpgradeHandler.CountUnProcessedPayload(), uint(5))
}

func TestLastKernelUpgradeHandler_ProcessLastKernelUpgradePayload_WrongTimeFormat(t *testing.T) {
	lastKernelUpgradeHandler := NewLastKernelUpgradeHandler()
	currentTime := time.Now().Format(time.RFC822Z)
	var payload = payload_handler.Payload{
		Type:    "last_kernel_upgrade",
		Payload: map[string]interface{}{"value": currentTime},
	}
	lastKernelUpgradeHandler.InsertPayload(payload)
	require.Error(t, lastKernelUpgradeHandler.ProcessLastKernelUpdatePayload())
}

func TestLastKernelUpgradeHandler_ProcessLastKernelUpgradePayload_InvalidData(t *testing.T) {
	lastKernelUpgradeHandler := NewLastKernelUpgradeHandler()
	var payload = payload_handler.Payload{
		Type:    "last_kernel_upgrade",
		Payload: map[string]interface{}{"value": "rubbish"},
	}
	lastKernelUpgradeHandler.InsertPayload(payload)
	require.Error(t, lastKernelUpgradeHandler.ProcessLastKernelUpdatePayload())
}

func TestLastKernelUpgradeHandler_ProcessLastKernelUpgradePayload(t *testing.T) {
	lastKernelUpgradeHandler := NewLastKernelUpgradeHandler()
	currentTime := time.Now().Format(time.RFC3339Nano)
	var payload = payload_handler.Payload{
		Type:    "last_kernel_upgrade",
		Payload: map[string]interface{}{"value": currentTime},
	}
	lastKernelUpgradeHandler.InsertPayload(payload)
	require.NoError(t, lastKernelUpgradeHandler.ProcessLastKernelUpdatePayload())
	require.Equal(t, lastKernelUpgradeHandler.latestUpgrade.Format(time.RFC3339Nano), currentTime)
}

func TestLastKernelUpgradeHandler_ProcessLastKernelUpgradePayload_Rolling_Latest_Upgrade(t *testing.T) {
	lastKernelUpgradeHandler := NewLastKernelUpgradeHandler()
	someTime := time.Now().Add(- 10 * time.Hour).Format(time.RFC3339Nano)
	var payload = payload_handler.Payload{
		Type:    "last_kernel_upgrade",
		Payload: map[string]interface{}{"value": someTime},
	}
	lastKernelUpgradeHandler.InsertPayload(payload)
	require.NoError(t, lastKernelUpgradeHandler.ProcessLastKernelUpdatePayload())
	require.Equal(t, lastKernelUpgradeHandler.latestUpgrade.Format(time.RFC3339Nano), someTime)


	//Payload should change if inserting a payload with a newer time
	currentTime := time.Now().Format(time.RFC3339Nano)
	var newPayload = payload_handler.Payload{
		Type:    "last_kernel_upgrade",
		Payload: map[string]interface{}{"value": currentTime},
	}

	lastKernelUpgradeHandler.InsertPayload(newPayload)
	require.NoError(t, lastKernelUpgradeHandler.ProcessLastKernelUpdatePayload())
	require.Equal(t, lastKernelUpgradeHandler.latestUpgrade.Format(time.RFC3339Nano), currentTime)

	//Payload should not change if inserting a payload with an older time
	anotherTime := time.Now().Add(- 5 * time.Hour).Format(time.RFC3339Nano)
	var anotherPayload = payload_handler.Payload{
		Type:    "last_kernel_upgrade",
		Payload: map[string]interface{}{"value": anotherTime},
	}

	lastKernelUpgradeHandler.InsertPayload(anotherPayload)
	require.NoError(t, lastKernelUpgradeHandler.ProcessLastKernelUpdatePayload())
	require.Equal(t, lastKernelUpgradeHandler.latestUpgrade.Format(time.RFC3339Nano), currentTime)
}

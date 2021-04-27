package dispatcher

import (
	"github.com/demoware/ingester/payload_handler"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	successResponseBody = `[
    {
		"type": "mock_data",
		"payload": {
            "value": 0.5481859
        }
	},
	{
		"type": "mock_data",
		"payload": {
            "value": 0.7747744
        }
	},
	{
		"type": "mock_data",
		"payload": {
            "value": 0.848484
        }
	}]`

	invalidResponseBody = `{
       "wrong_body":"bbdgdgd"
     }
	 `
)

func TestDispatcher_GetHandler(t *testing.T) {
	var payloads []payload_handler.Payload
	options := DispatcherOptions{
		BatchLimit: 1,
	}
	dispatcher := options.NewDispatcher()
	dispatcher.NewHandler(&payload_handler.MockPayloadHandler{
		Payload: payloads,
	})
	require.NotNil(t, dispatcher.GetHandler("mock_data"))

}
func TestDispatcher_CanProcessNewPayload(t *testing.T) {
	var payloads []payload_handler.Payload
	payloads = append(payloads, payload_handler.Payload{
		Type:    "mock_data",
		Payload: "data1",
	}, payload_handler.Payload{
		Type:    "mock_data",
		Payload: "data2",
	}, payload_handler.Payload{
		Type:    "mock_data",
		Payload: "data3",
	})
	options := DispatcherOptions{
		BatchLimit: 2,
	}
	dispatcher := options.NewDispatcher()
	mockHandler := &payload_handler.MockPayloadHandler{
		Payload: payloads,
	}
	dispatcher.NewHandler(mockHandler)

	require.Equal(t, dispatcher.CanProcessNewPayload(), false)

	mockHandler.Payload = payloads[0:1]

	require.Equal(t, dispatcher.CanProcessNewPayload(), true)

}

func TestDispatcher_ProcessPayload(t *testing.T) {
	oas := httptest.NewServer(getSuccessfulResponse(t, successResponseBody))
	defer oas.Close()

	var payloads []payload_handler.Payload
	options := DispatcherOptions{
		BatchLimit: 5,
		AuthToken:  "testtoken",
		MetricsUri: oas.URL,
	}
	dispatcher := options.NewDispatcher()
	mockHandler := &payload_handler.MockPayloadHandler{
		Payload: payloads,
	}
	dispatcher.NewHandler(mockHandler)

	require.NoError(t, dispatcher.ProcessPayload())
	require.Equal(t, len(mockHandler.Payload), 3)
}

func TestDispatcher_ProcessPayload_Error(t *testing.T) {
	oas := httptest.NewServer(errorResponse(t))
	defer oas.Close()

	var payloads []payload_handler.Payload
	options := DispatcherOptions{
		BatchLimit: 5,
		AuthToken:  "testtoken",
		MetricsUri: oas.URL,
	}
	dispatcher := options.NewDispatcher()
	mockHandler := &payload_handler.MockPayloadHandler{
		Payload: payloads,
	}
	dispatcher.NewHandler(mockHandler)

	require.Error(t, dispatcher.ProcessPayload())
}

func TestDispatcher_ProcessPayload_Error_InvalidBody(t *testing.T) {
	oas := httptest.NewServer(getSuccessfulResponse(t, invalidResponseBody))
	defer oas.Close()

	var payloads []payload_handler.Payload
	options := DispatcherOptions{
		BatchLimit: 5,
		AuthToken:  "testtoken",
		MetricsUri: oas.URL,
	}
	dispatcher := options.NewDispatcher()
	mockHandler := &payload_handler.MockPayloadHandler{
		Payload: payloads,
	}
	dispatcher.NewHandler(mockHandler)

	require.Error(t, dispatcher.ProcessPayload())
}

func getSuccessfulResponse(t *testing.T, body string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		require.Equal(t, "Basic dGVzdHRva2VuOg==", req.Header.Get("Authorization"))

		rw.Header().Set("Content-Type", "application/json")
		_, err := rw.Write([]byte(body))

		require.NoError(t, err)
	}
}

func errorResponse(t *testing.T) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		require.Equal(t, "Basic dGVzdHRva2VuOg==", req.Header.Get("Authorization"))
		rw.WriteHeader(503)
	}
}

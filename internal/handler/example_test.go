package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/danilov-go/metrics-alerting.git/internal/handler"
	"github.com/danilov-go/metrics-alerting.git/internal/logger"
	"github.com/danilov-go/metrics-alerting.git/internal/repository"
)

func ExampleMetricsHandler_ApiUpdateHandler() {
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	config := repository.ConfigFile{}
	storage := repository.InitMemStorage(config, logger.Log.Sugar())
	h := handler.NewMetricsHandler(storage, logger.Log.Sugar())
	metric := `{"id":"Alloc","type":"gauge","value":123.45}`
	req := httptest.NewRequest(http.MethodPost, "/update/", strings.NewReader(metric))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ApiUpdateHandler().ServeHTTP(w, req)
	fmt.Println(w.Code)
	fmt.Println(w.Header().Get("Content-Type"))

	// Output:
	// 200
	// application/json
}

func ExampleMetricsHandler_ApiValueHandler() {
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	config := repository.ConfigFile{}
	storage := repository.InitMemStorage(config, logger.Log.Sugar())
	h := handler.NewMetricsHandler(storage, logger.Log.Sugar())
	var deltaValue int64 = 5
	_ = storage.SaveCounters(context.Background(), "PollCount", deltaValue)
	metric := `{"id":"PollCount","type":"counter"}`
	req := httptest.NewRequest(http.MethodPost, "/value/", strings.NewReader(metric))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ApiValueHandler().ServeHTTP(w, req)
	fmt.Println(w.Code)
	fmt.Println(w.Header().Get("Content-Type"))
	fmt.Println(w.Body.String())

	// Output:
	// 200
	// application/json
	// {"id":"PollCount","type":"counter","delta":5}
}

func ExampleMetricsHandler_ApiUpdatesHandler() {
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	config := repository.ConfigFile{}
	storage := repository.InitMemStorage(config, logger.Log.Sugar())
	h := handler.NewMetricsHandler(storage, logger.Log.Sugar())
	metrics := `[
		{"id":"Alloc","type":"gauge","value":123.45},
		{"id":"PollCount","type":"counter","delta":5}
	]`
	req := httptest.NewRequest(http.MethodPost, "/updates/", strings.NewReader(metrics))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ApiUpdatesHandler().ServeHTTP(w, req)
	fmt.Println(w.Code)
	fmt.Println(w.Header().Get("Content-Type"))

	// Output:
	// 200
	// application/json
}

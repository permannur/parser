package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"parser/benchmark"
	"parser/config"
	"parser/logger"
	"parser/yandex"
	"time"
)

func Search(w http.ResponseWriter, r *http.Request) {
	// create a context with configs context timeout
	ctx, cancel := context.WithTimeout(r.Context(), config.Values().GetContextTimeout()*time.Second)
	defer cancel()

	// get response from yandex
	res, err := yandex.Get(ctx, url.QueryEscape(r.FormValue("search")))
	if err != nil {
		logger.Write(fmt.Sprintf("handler.Search: err=%s\n", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if res.Error != nil {
		logger.Write(fmt.Sprintf("handler.Search: res.Error=%s\n", res.Error))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// run benchmark and get json in bytes
	var jsonBt []byte
	jsonBt, err = benchmark.GetRecommendNums(ctx, res.Items)
	if err != nil {
		logger.Write(fmt.Sprintf("handler.Search: err=%s\n", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// write to responseWriter
	_, err = w.Write(jsonBt)
	if err != nil {
		logger.Write(fmt.Sprintf("handler.Search: err=%s\n", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

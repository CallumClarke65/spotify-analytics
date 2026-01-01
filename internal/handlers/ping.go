package handlers

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

// PingResponse godoc
// @Description Response for ping endpoint
// @name PingResponse
type PingResponse struct {
	Time     string           `json:"time"`
	MemStats MemStatsResponse `json:"mem_stats"`
	Status   string           `json:"status"`
}

// MemStatsResponse godoc
// @Description Go runtime memory stats
// @name MemStatsResponse
type MemStatsResponse struct {
	Alloc      uint64 `json:"alloc_bytes"`
	TotalAlloc uint64 `json:"total_alloc_bytes"`
	Sys        uint64 `json:"sys_bytes"`
	HeapAlloc  uint64 `json:"heap_alloc_bytes"`
	HeapSys    uint64 `json:"heap_sys_bytes"`
	NumGC      uint32 `json:"num_gc"`
}

// Ping godoc
// @Summary Ping / health check
// @Description Returns service health and Go runtime memory statistics
// @Tags health
// @Produce json
// @Success 200 {object} PingResponse
// @Router /ping [get]
func Ping(w http.ResponseWriter, r *http.Request) {

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	mem := MemStatsResponse{
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		Sys:        m.Sys,
		HeapAlloc:  m.HeapAlloc,
		HeapSys:    m.HeapSys,
		NumGC:      m.NumGC,
	}

	resp := PingResponse{
		Time:     time.Now().Format(time.RFC3339),
		Status:   "ok",
		MemStats: mem,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

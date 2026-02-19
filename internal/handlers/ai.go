package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"biviz/internal/ai"
	"biviz/internal/middleware"
)

const systemPrompt = "You are BiViz AI Assistant — a helpful data analysis and BI (Business Intelligence) expert. " +
	"Answer concisely in Korean. Use markdown formatting when helpful. " +
	"Help users understand data, suggest chart types, write SQL queries, and interpret analytics."

type chatRequest struct {
	Message string            `json:"message"`
	History []ai.HistoryEntry `json:"history"`
}

// HandleAIChat POST /api/ai/chat — SSE 스트리밍 AI 채팅
func HandleAIChat(w http.ResponseWriter, r *http.Request) {
	session := middleware.GetCurrentSession(r)
	if session == nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if !ai.IsAvailable() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, `{"error":"AI service is not configured"}`)
		return
	}

	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"Invalid request body"}`)
		return
	}

	if req.Message == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"Message is required"}`)
		return
	}

	history := []ai.HistoryEntry{
		{Role: "system", Content: systemPrompt},
	}
	history = append(history, req.History...)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	chunks, errc := ai.ChatStream(r.Context(), history, req.Message)

	for chunk := range chunks {
		data, _ := json.Marshal(map[string]string{"text": chunk})
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	if err := <-errc; err != nil {
		data, _ := json.Marshal(map[string]string{"error": err.Error()})
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
		return
	}

	fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
}

package relay

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestEstimateTokens(t *testing.T) {
	cases := []struct {
		name string
		body string
		want int
	}{
		{"empty object", `{}`, 0},
		{"numbers and bools ignored", `{"max_tokens":1024,"stream":false,"thinking":{"budget_tokens":2048}}`, 0},
		{"text message", `{"model":"m","messages":[{"role":"user","content":"aaaaaaaaaaaaaaaaaaa"}]}`, 9},
		{"system block array", `{"system":[{"type":"text","text":"hi"}]}`, 3},
		{"tool schema", `{"tools":[{"name":"get_weather","description":"Get weather","input_schema":{"type":"object"}}]}`, 10},
		{"image base64 not counted", `{"messages":[{"role":"user","content":[{"type":"image","source":{"data":"` + strings.Repeat("A", 100000) + `"}}]}]}`, 1602},
		{"document base64 not counted", `{"messages":[{"role":"user","content":[{"type":"document","source":{"data":"` + strings.Repeat("A", 100000) + `"}}]}]}`, 1602},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var payload any
			if err := json.Unmarshal([]byte(tc.body), &payload); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got := estimateTokens(payload); got != tc.want {
				t.Errorf("estimateTokens = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestCountTokensHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.POST("/v1/messages/count_tokens", CountTokens)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/messages/count_tokens",
		strings.NewReader(`{"messages":[{"role":"user","content":"hello"}]}`))
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		InputTokens int `json:"input_tokens"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.InputTokens != 4 {
		t.Errorf("input_tokens = %d, want 4", response.InputTokens)
	}

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/v1/messages/count_tokens", strings.NewReader("not json"))
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("invalid body status = %d, want 400", recorder.Code)
	}
}

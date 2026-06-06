package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Arturikou/urlshortener/internal/config"
	"github.com/Arturikou/urlshortener/internal/handlers"
	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/Arturikou/urlshortener/internal/service/shortener"
	"github.com/Arturikou/urlshortener/internal/userctx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	exampleCfg    = config.HandlersConfig{BaseAddr: "http://localhost:8080"}
	exampleUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
)

// --- stubs ---

type stubURLService struct{}

func (s *stubURLService) AddURL(_ context.Context, _ string) (shortener.AddURLResult, error) {
	return shortener.AddURLResult{Alias: "EwHXdJfB", IsInsert: true}, nil
}

func (s *stubURLService) GetURL(_ context.Context, _ string) (string, error) {
	return "https://practicum.yandex.ru", nil
}

func (s *stubURLService) AddURLs(_ context.Context, batch []models.ShortenBatch) ([]models.ShortenBatch, error) {
	for i := range batch {
		batch[i].Alias = fmt.Sprintf("alias%d", i+1)
	}
	return batch, nil
}

func (s *stubURLService) GetUserURLs(_ context.Context, _ uuid.UUID) ([]models.UserUrls, error) {
	return []models.UserUrls{
		{OriginalURL: "https://practicum.yandex.ru", Alias: "EwHXdJfB"},
	}, nil
}

func (s *stubURLService) GetStats(_ context.Context) (models.Stats, error) {
	return models.Stats{}, nil
}

type stubDeleteEnqueuer struct{}

func (e *stubDeleteEnqueuer) EnqueueDelete(_ uuid.UUID, _ []string) {}

type stubPinger struct{}

func (p *stubPinger) Ping(_ context.Context) error { return nil }

func newStubHandlers() *handlers.Handlers {
	return handlers.New(&stubURLService{}, &stubDeleteEnqueuer{}, &stubPinger{}, exampleCfg, zap.NewNop().Sugar())
}

// withUserID injects a user UUID into the request context.
func withUserID(req *http.Request, id uuid.UUID) *http.Request {
	ctx := userctx.WithUserID(req.Context(), id)
	return req.WithContext(ctx)
}

// --- examples ---

func ExampleHandlers_Ping() {
	h := newStubHandlers()
	r := chi.NewRouter()
	r.Get("/ping", h.Ping)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println("Status:", w.Result().Status)
	// Output:
	// Status: 200 OK
}

func ExampleHandlers_AddURL() {
	h := newStubHandlers()
	r := chi.NewRouter()
	r.Post("/", h.AddURL)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println("Status:", w.Result().Status)
	fmt.Println("Short URL:", strings.TrimSpace(w.Body.String()))
	// Output:
	// Status: 201 Created
	// Short URL: http://localhost:8080/EwHXdJfB
}

func ExampleHandlers_GetURL() {
	h := newStubHandlers()
	r := chi.NewRouter()
	r.Get("/{alias}", h.GetURL)

	req := httptest.NewRequest(http.MethodGet, "/EwHXdJfB", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println("Status:", w.Result().Status)
	fmt.Println("Location:", w.Header().Get("Location"))
	// Output:
	// Status: 307 Temporary Redirect
	// Location: https://practicum.yandex.ru
}

func ExampleHandlers_Shorten() {
	h := newStubHandlers()
	r := chi.NewRouter()
	r.Post("/api/shorten", h.Shorten)

	body, _ := json.Marshal(handlers.ShortenReq{URL: "https://practicum.yandex.ru"})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println("Status:", w.Result().Status)
	var resp handlers.ShortenResp
	_ = json.NewDecoder(w.Body).Decode(&resp)
	fmt.Println("Result:", resp.Result)
	// Output:
	// Status: 201 Created
	// Result: http://localhost:8080/EwHXdJfB
}

func ExampleHandlers_ShortenBatch() {
	h := newStubHandlers()
	r := chi.NewRouter()
	r.Post("/api/shorten/batch", h.ShortenBatch)

	reqBody := []handlers.ShortenBatchReq{
		{CorrelationID: "1", OriginalURL: "https://practicum.yandex.ru"},
		{CorrelationID: "2", OriginalURL: "https://google.com"},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println("Status:", w.Result().Status)
	var resp []handlers.ShortenBatchResp
	_ = json.NewDecoder(w.Body).Decode(&resp)
	for _, item := range resp {
		fmt.Printf("correlation_id=%s short_url=%s\n", item.CorrelationID, item.ShortURL)
	}
	// Output:
	// Status: 201 Created
	// correlation_id=1 short_url=http://localhost:8080/alias1
	// correlation_id=2 short_url=http://localhost:8080/alias2
}

func ExampleHandlers_UserUrls() {
	h := newStubHandlers()
	r := chi.NewRouter()
	r.Get("/api/user/urls", h.UserUrls)

	req := withUserID(httptest.NewRequest(http.MethodGet, "/api/user/urls", nil), exampleUserID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println("Status:", w.Result().Status)
	var resp []handlers.UserUrlsResp
	_ = json.NewDecoder(w.Body).Decode(&resp)
	for _, item := range resp {
		fmt.Printf("short_url=%s original_url=%s\n", item.ShortURL, item.OriginalURL)
	}
	// Output:
	// Status: 200 OK
	// short_url=http://localhost:8080/EwHXdJfB original_url=https://practicum.yandex.ru
}

func ExampleHandlers_DeleteUserURLs() {
	h := newStubHandlers()
	r := chi.NewRouter()
	r.Delete("/api/user/urls", h.DeleteUserURLs)

	body, _ := json.Marshal([]string{"EwHXdJfB", "AbCdEfGh"})
	req := withUserID(
		httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body)),
		exampleUserID,
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println("Status:", w.Result().Status)
	// Output:
	// Status: 202 Accepted
}

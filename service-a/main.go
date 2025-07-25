package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.21.0"
)

type CEPRequest struct {
	CEP string `json:"cep"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func initTracer() (*trace.TracerProvider, error) {
	ctx := context.Background()

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("service-a"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp, nil
}

func validateCEP(cep string) bool {
	// Verifica se tem exatamente 8 dígitos
	if len(cep) != 8 {
		return false
	}

	// Verifica se contém apenas números
	matched, _ := regexp.MatchString(`^\d{8}$`, cep)
	return matched
}

func handleCEP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("service-a")

	ctx, span := tracer.Start(ctx, "handle-cep-request")
	defer span.End()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CEPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "invalid request body"})
		return
	}

	// Validar CEP
	ctx, validateSpan := tracer.Start(ctx, "validate-cep")
	if !validateCEP(req.CEP) {
		validateSpan.RecordError(fmt.Errorf("invalid zipcode: %s", req.CEP))
		validateSpan.End()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "invalid zipcode"})
		return
	}
	validateSpan.End()

	// Chamar Serviço B
	ctx, callServiceBSpan := tracer.Start(ctx, "call-service-b")
	defer callServiceBSpan.End()

	serviceBURL := os.Getenv("SERVICE_B_URL")
	if serviceBURL == "" {
		serviceBURL = "http://localhost:8081"
	}

	reqBody, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", serviceBURL+"/weather", bytes.NewBuffer(reqBody))
	if err != nil {
		callServiceBSpan.RecordError(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Propagar contexto de tracing
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(httpReq.Header))

	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   30 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		callServiceBSpan.RecordError(err)
		http.Error(w, "Failed to call service B", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Repassar resposta do Serviço B
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)

	var responseBody interface{}
	json.NewDecoder(resp.Body).Decode(&responseBody)
	json.NewEncoder(w).Encode(responseBody)
}

func main() {
	tp, err := initTracer()
	if err != nil {
		log.Fatal("Failed to initialize tracer:", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleCEP)

	handler := otelhttp.NewHandler(mux, "service-a")

	fmt.Println("Service A starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

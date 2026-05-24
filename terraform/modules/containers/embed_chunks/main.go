package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ── Clientes AWS ──────────────────────────────────────────────────────────────

var (
	s3Client  *s3.Client
	ddbClient *dynamodb.Client
	bucket    string
	execTable string
	ollamaURL string
)

// ── Tipos ─────────────────────────────────────────────────────────────────────

type Chunk struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Position int    `json:"position"`
	Tokens   int    `json:"tokens"`
}

type ChunkWithVector struct {
	Chunk
	Vector []float64 `json:"vector"`
	Model  string    `json:"model"`
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type OllamaResponse struct {
	Embedding []float64 `json:"embedding"`
}

type TaskInput struct {
	TempChunksKey string `json:"temp_chunks_key"`
	Bucket        string `json:"bucket"`
	ExecutionID   string `json:"execution_id"`
	EmbedModel    string `json:"embed_model"`
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("[embed-chunks] failed to load AWS config: %v", err)
	}

	s3Client  = s3.NewFromConfig(cfg)
	ddbClient = dynamodb.NewFromConfig(cfg)
	bucket    = mustEnv("DOCUMENTS_BUCKET")
	execTable = mustEnv("EXECUTIONS_TABLE")
	ollamaURL = getEnv("OLLAMA_URL", "http://localhost:11434")

	// Lê input via variável de ambiente (Step Function passa via env)
	inputJSON := mustEnv("TASK_INPUT")
	var input TaskInput
	if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
		log.Fatalf("[embed-chunks] failed to parse TASK_INPUT: %v", err)
	}

	log.Printf("[embed-chunks] start — execution_id=%s model=%s", input.ExecutionID, input.EmbedModel)

	if err := run(ctx, input); err != nil {
		updateStatus(ctx, input.ExecutionID, "embed_chunks", "error", 55, err.Error())
		log.Fatalf("[embed-chunks] fatal error: %v", err)
	}
}

func run(ctx context.Context, input TaskInput) error {
	updateStatus(ctx, input.ExecutionID, "embed_chunks", "running", 52, "Aguardando Ollama...")

	// Aguarda Ollama estar pronto
	if err := waitOllama(); err != nil {
		return fmt.Errorf("ollama not ready: %w", err)
	}

	updateStatus(ctx, input.ExecutionID, "embed_chunks", "running", 55, "Baixando chunks do S3...")

	// Baixa chunks.json do S3 temp
	result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(input.Bucket),
		Key:    aws.String(input.TempChunksKey),
	})
	if err != nil {
		return fmt.Errorf("failed to download chunks: %w", err)
	}
	defer result.Body.Close()

	var chunks []Chunk
	if err := json.NewDecoder(result.Body).Decode(&chunks); err != nil {
		return fmt.Errorf("failed to parse chunks: %w", err)
	}

	log.Printf("[embed-chunks] loaded %d chunks", len(chunks))

	// Embeda cada chunk
	model := input.EmbedModel
	if model == "" {
		model = "nomic-embed-text"
	}

	chunksWithVectors := make([]ChunkWithVector, 0, len(chunks))
	for i, chunk := range chunks {
		progress := 55 + int(float64(i)/float64(len(chunks))*25)
		if i%10 == 0 {
			msg := fmt.Sprintf("Embedding chunk %d/%d...", i+1, len(chunks))
			log.Printf("[embed-chunks] %s", msg)
			updateStatus(ctx, input.ExecutionID, "embed_chunks", "running", progress, msg)
		}

		vector, err := embed(chunk.Text, model)
		if err != nil {
			return fmt.Errorf("failed to embed chunk %s: %w", chunk.ID, err)
		}

		chunksWithVectors = append(chunksWithVectors, ChunkWithVector{
			Chunk:  chunk,
			Vector: vector,
			Model:  model,
		})
	}

	log.Printf("[embed-chunks] all %d chunks embedded", len(chunksWithVectors))
	updateStatus(ctx, input.ExecutionID, "embed_chunks", "running", 80, "Salvando vetores no S3...")

	// Salva resultado no S3 temp
	tempVectorsKey := fmt.Sprintf("temp/%s/vectors.json", input.ExecutionID)
	data, err := json.Marshal(chunksWithVectors)
	if err != nil {
		return fmt.Errorf("failed to marshal vectors: %w", err)
	}

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(tempVectorsKey),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload vectors: %w", err)
	}

	log.Printf("[embed-chunks] vectors saved to s3://%s/%s", bucket, tempVectorsKey)
	updateStatus(ctx, input.ExecutionID, "embed_chunks", "done", 82,
		fmt.Sprintf("%d chunks embedados com %s", len(chunksWithVectors), model))

	return nil
}

// ── Ollama ────────────────────────────────────────────────────────────────────

func waitOllama() error {
	for i := 1; i <= 20; i++ {
		resp, err := http.Get(ollamaURL)
		if err == nil && resp.StatusCode == 200 {
			log.Printf("[embed-chunks] Ollama ready after %d attempts", i)
			return nil
		}
		log.Printf("[embed-chunks] waiting for Ollama attempt %d/20...", i)
		time.Sleep(3 * time.Second)
	}
	return fmt.Errorf("ollama did not respond after 60s")
}

func embed(text, model string) ([]float64, error) {
	reqBody, _ := json.Marshal(OllamaRequest{
		Model:  model,
		Prompt: text,
	})

	resp, err := http.Post(
		ollamaURL+"/api/embeddings",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to parse ollama response: %w", err)
	}

	return ollamaResp.Embedding, nil
}

// ── DynamoDB ──────────────────────────────────────────────────────────────────

func updateStatus(ctx context.Context, executionID, step, status string, progress int, message string) {
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	now := time.Now().In(loc).Format("02/01/2006 15:04:05")

	_, err := ddbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(execTable),
		Item: map[string]types.AttributeValue{
			"execution_id": &types.AttributeValueMemberS{Value: executionID},
			"step":         &types.AttributeValueMemberS{Value: step},
			"status":       &types.AttributeValueMemberS{Value: status},
			"progress":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", progress)},
			"message":      &types.AttributeValueMemberS{Value: message},
			"updated_at":   &types.AttributeValueMemberS{Value: now},
		},
	})
	if err != nil {
		log.Printf("[embed-chunks] failed to update DynamoDB: %v", err)
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env var: %s", key)
	}
	return v
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

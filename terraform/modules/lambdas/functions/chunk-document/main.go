package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
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
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}
	s3Client  = s3.NewFromConfig(cfg)
	ddbClient = dynamodb.NewFromConfig(cfg)
	bucket    = mustEnv("DOCUMENTS_BUCKET")
	execTable = mustEnv("EXECUTIONS_TABLE")
}

// ── Tipos ─────────────────────────────────────────────────────────────────────

type ChunkingParams struct {
	ChunkSize int `json:"chunk_size"`
	Overlap   int `json:"overlap"`
}

type ChunkingConfig struct {
	Method string         `json:"method"`
	Params ChunkingParams `json:"params"`
}

type PayloadData struct {
	Chunking ChunkingConfig `json:"chunking"`
}

type Chunk struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Position int    `json:"position"`
	Tokens   int    `json:"tokens"`
}

type Input struct {
	S3Key       string          `json:"s3_key"`
	Bucket      string          `json:"bucket"`
	TempTextKey string          `json:"temp_text_key"`
	ExecutionID string          `json:"execution_id"`
	Payload     json.RawMessage `json:"payload"`
}

type Output struct {
	S3Key         string          `json:"s3_key"`
	Bucket        string          `json:"bucket"`
	TempChunksKey string          `json:"temp_chunks_key"`
	ChunkCount    int             `json:"chunk_count"`
	ExecutionID   string          `json:"execution_id"`
	Payload       json.RawMessage `json:"payload"`
}

// ── Handler ───────────────────────────────────────────────────────────────────

func handler(ctx context.Context, input Input) (Output, error) {
	updateStatus(ctx, input.ExecutionID, "chunk_document", "running", 30, "Baixando texto do S3...")

	// Baixa o texto extraído do S3 temp
	result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(input.Bucket),
		Key:    aws.String(input.TempTextKey),
	})
	if err != nil {
		updateStatus(ctx, input.ExecutionID, "chunk_document", "error", 30, err.Error())
		return Output{}, fmt.Errorf("failed to download text: %w", err)
	}
	defer result.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)
	text := buf.String()

	// Lê configuração de chunking do payload
	var payload PayloadData
	if err := json.Unmarshal(input.Payload, &payload); err != nil {
		updateStatus(ctx, input.ExecutionID, "chunk_document", "error", 30, err.Error())
		return Output{}, fmt.Errorf("failed to parse payload: %w", err)
	}

	updateStatus(ctx, input.ExecutionID, "chunk_document", "running", 35, "Dividindo documento em chunks...")

	// Aplica chunking conforme strategy
	var chunks []Chunk
	switch payload.Chunking.Method {
	case "fixed_size":
		chunks = chunkFixedSize(text, payload.Chunking.Params.ChunkSize, payload.Chunking.Params.Overlap)
	case "sentence":
		chunks = chunkBySentence(text, payload.Chunking.Params.ChunkSize)
	case "paragraph":
		chunks = chunkByParagraph(text)
	default:
		chunks = chunkFixedSize(text, 512, 64)
	}

	updateStatus(ctx, input.ExecutionID, "chunk_document", "running", 45,
		fmt.Sprintf("Salvando %d chunks no S3...", len(chunks)))

	// Serializa chunks e salva no S3 temp
	chunksJSON, err := json.Marshal(chunks)
	if err != nil {
		updateStatus(ctx, input.ExecutionID, "chunk_document", "error", 45, err.Error())
		return Output{}, fmt.Errorf("failed to marshal chunks: %w", err)
	}

	tempChunksKey := fmt.Sprintf("temp/%s/chunks.json", input.ExecutionID)
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(tempChunksKey),
		Body:        bytes.NewReader(chunksJSON),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		updateStatus(ctx, input.ExecutionID, "chunk_document", "error", 45, err.Error())
		return Output{}, fmt.Errorf("failed to upload chunks: %w", err)
	}

	updateStatus(ctx, input.ExecutionID, "chunk_document", "done", 50,
		fmt.Sprintf("%d chunks gerados", len(chunks)))

	return Output{
		S3Key:         input.S3Key,
		Bucket:        input.Bucket,
		TempChunksKey: tempChunksKey,
		ChunkCount:    len(chunks),
		ExecutionID:   input.ExecutionID,
		Payload:       input.Payload,
	}, nil
}

// ── Chunking strategies ───────────────────────────────────────────────────────

func chunkFixedSize(text string, size, overlap int) []Chunk {
	words  := strings.Fields(text)
	chunks := []Chunk{}

	i := 0
	pos := 0
	for i < len(words) {
		end := i + size
		if end > len(words) {
			end = len(words)
		}
		chunkText := strings.Join(words[i:end], " ")
		chunks = append(chunks, Chunk{
			ID:       fmt.Sprintf("chunk_%04d", pos),
			Text:     chunkText,
			Position: pos,
			Tokens:   end - i,
		})
		pos++
		i += size - overlap
		if i < 0 {
			i = 0
		}
	}
	return chunks
}

func chunkBySentence(text string, sentencesPerChunk int) []Chunk {
	sentences := strings.Split(text, ". ")
	chunks    := []Chunk{}

	for i := 0; i < len(sentences); i += sentencesPerChunk {
		end := i + sentencesPerChunk
		if end > len(sentences) {
			end = len(sentences)
		}
		chunkText := strings.Join(sentences[i:end], ". ")
		chunks = append(chunks, Chunk{
			ID:       fmt.Sprintf("chunk_%04d", i/sentencesPerChunk),
			Text:     chunkText,
			Position: i / sentencesPerChunk,
			Tokens:   len(strings.Fields(chunkText)),
		})
	}
	return chunks
}

func chunkByParagraph(text string) []Chunk {
	paragraphs := strings.Split(text, "\n\n")
	chunks     := []Chunk{}

	for i, p := range paragraphs {
		p = strings.TrimSpace(p)
		if len(p) < 50 {
			continue
		}
		chunks = append(chunks, Chunk{
			ID:       fmt.Sprintf("chunk_%04d", i),
			Text:     p,
			Position: i,
			Tokens:   len(strings.Fields(p)),
		})
	}
	return chunks
}

// ── DynamoDB status ───────────────────────────────────────────────────────────

func updateStatus(ctx context.Context, executionID, step, status string, progress int, message string) {
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	now := time.Now().In(loc).Format("02/01/2006 15:04:05")

	ddbClient.PutItem(ctx, &dynamodb.PutItemInput{
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
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("missing required env var: %s", key))
	}
	return v
}

func main() {
	lambda.Start(handler)
}

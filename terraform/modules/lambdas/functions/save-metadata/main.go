package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

var (
	s3Client       *s3.Client
	ddbClient      *dynamodb.Client
	signer         *v4.Signer
	awsCfg         aws.Config
	bucket         string
	execTable      string
	documentsTable string
	chunksTable    string
	vectorBucket   string
	vectorIndex    string
	awsRegion      string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}
	awsCfg = cfg
	s3Client = s3.NewFromConfig(cfg)
	ddbClient = dynamodb.NewFromConfig(cfg)
	signer = v4.NewSigner()
	bucket = mustEnv("DOCUMENTS_BUCKET")
	execTable = mustEnv("EXECUTIONS_TABLE")
	documentsTable = mustEnv("DOCUMENTS_TABLE")
	chunksTable = mustEnv("CHUNKS_TABLE")
	vectorBucket = mustEnv("VECTOR_BUCKET")
	vectorIndex = mustEnv("VECTOR_INDEX")
	awsRegion = getEnv("AWS_REGION", "us-east-1")
	log.Printf("[save-metadata] init OK")
}

type ChunkWithVector struct {
	ID       string    `json:"id"`
	Text     string    `json:"text"`
	Position int       `json:"position"`
	Tokens   int       `json:"tokens"`
	Vector   []float64 `json:"vector"`
	Model    string    `json:"model"`
}

type PayloadMetadata struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Language    string   `json:"language"`
	DocType     string   `json:"doc_type"`
	Tags        []string `json:"tags"`
	Source      string   `json:"source"`
	Visibility  string   `json:"visibility"`
}

type PayloadDocument struct {
	Filename string `json:"filename"`
	S3Key    string `json:"s3_key"`
}

type PayloadEmbedding struct {
	Model string `json:"model"`
}

type PayloadChunking struct {
	Method string `json:"method"`
}

type Payload struct {
	Document   PayloadDocument  `json:"document"`
	Embedding  PayloadEmbedding `json:"embedding"`
	Chunking   PayloadChunking  `json:"chunking"`
	Metadata   PayloadMetadata  `json:"metadata"`
	DeployedAt string           `json:"deployed_at"`
}

type Input struct {
	S3Key       string          `json:"s3_key"`
	Bucket      string          `json:"bucket"`
	ExecutionID string          `json:"execution_id"`
	ChunkCount  int             `json:"chunk_count"`
	Payload     json.RawMessage `json:"payload"`
}

// S3 Vectors REST API types
type S3VectorItem struct {
	Key      string                 `json:"key"`
	Data     S3VectorData           `json:"data"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type S3VectorData struct {
	Float32 []float32 `json:"float32"`
}

type PutVectorsBody struct {
	VectorBucketName string         `json:"vectorBucketName"`
	IndexName        string         `json:"indexName"`
	Vectors          []S3VectorItem `json:"vectors"`
}

func handler(ctx context.Context, input Input) (map[string]interface{}, error) {
	log.Printf("[save-metadata] start — execution_id=%s", input.ExecutionID)
	updateStatus(ctx, input.ExecutionID, "save_metadata", "running", 83, "Lendo vetores do S3...")

	var payload Payload
	if err := json.Unmarshal(input.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse payload: %w", err)
	}

	// Baixa vectors.json
	tempVectorsKey := fmt.Sprintf("temp/%s/vectors.json", input.ExecutionID)
	result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(input.Bucket),
		Key:    aws.String(tempVectorsKey),
	})
	if err != nil {
		updateStatus(ctx, input.ExecutionID, "save_metadata", "error", 83, err.Error())
		return nil, fmt.Errorf("failed to download vectors: %w", err)
	}
	defer result.Body.Close()

	var chunks []ChunkWithVector
	if err := json.NewDecoder(result.Body).Decode(&chunks); err != nil {
		updateStatus(ctx, input.ExecutionID, "save_metadata", "error", 83, err.Error())
		return nil, fmt.Errorf("failed to parse vectors: %w", err)
	}

	log.Printf("[save-metadata] loaded %d chunks", len(chunks))
	documentID := uuid.New().String()
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	now := time.Now().In(loc).Format("02/01/2006 15:04:05")

	// Determina o index pelo modelo
	if len(chunks) > 0 {
		vectorIndex = indexForModel(chunks[0].Model)
		log.Printf("[save-metadata] using index %s for model %s", vectorIndex, chunks[0].Model)
	}

	// 1. Salva chunks no DynamoDB
	updateStatus(ctx, input.ExecutionID, "save_metadata", "running", 85, "Salvando chunks no DynamoDB...")
	for _, chunk := range chunks {
		tags, _ := json.Marshal(payload.Metadata.Tags)
		ddbClient.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(chunksTable),
			Item: map[string]types.AttributeValue{
				"chunk_id":    &types.AttributeValueMemberS{Value: chunk.ID},
				"document_id": &types.AttributeValueMemberS{Value: documentID},
				"text":        &types.AttributeValueMemberS{Value: chunk.Text},
				"position":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", chunk.Position)},
				"tokens":      &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", chunk.Tokens)},
				"model":       &types.AttributeValueMemberS{Value: chunk.Model},
				"created_at":  &types.AttributeValueMemberS{Value: now},
				"tags":        &types.AttributeValueMemberS{Value: string(tags)},
			},
		})
	}
	log.Printf("[save-metadata] %d chunks saved to DynamoDB", len(chunks))

	// 2. Salva vetores no S3 Vectors via REST API
	updateStatus(ctx, input.ExecutionID, "save_metadata", "running", 90, "Salvando vetores no S3 Vectors...")

	batchSize := 100
	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		batch := chunks[i:end]

		items := make([]S3VectorItem, 0, len(batch))
		for _, chunk := range batch {
			vec32 := make([]float32, len(chunk.Vector))
			for j, v := range chunk.Vector {
				vec32[j] = float32(v)
			}
			items = append(items, S3VectorItem{
				Key:  chunk.ID,
				Data: S3VectorData{Float32: vec32},
				Metadata: map[string]interface{}{
					"document_id": documentID,
					"chunk_id":    chunk.ID,
					"model":       chunk.Model,
				},
			})
		}

		if err := putVectorsREST(ctx, items); err != nil {
			log.Printf("[save-metadata] warning: S3 Vectors batch %d failed: %v", end, err)
		} else {
			log.Printf("[save-metadata] saved vectors batch %d/%d", end, len(chunks))
		}
	}

	// 3. Salva documento
	updateStatus(ctx, input.ExecutionID, "save_metadata", "running", 95, "Salvando documento...")
	tags, _ := json.Marshal(payload.Metadata.Tags)
	ddbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(documentsTable),
		Item: map[string]types.AttributeValue{
			"document_id":  &types.AttributeValueMemberS{Value: documentID},
			"filename":     &types.AttributeValueMemberS{Value: payload.Document.Filename},
			"s3_key":       &types.AttributeValueMemberS{Value: payload.Document.S3Key},
			"title":        &types.AttributeValueMemberS{Value: payload.Metadata.Title},
			"description":  &types.AttributeValueMemberS{Value: payload.Metadata.Description},
			"language":     &types.AttributeValueMemberS{Value: payload.Metadata.Language},
			"doc_type":     &types.AttributeValueMemberS{Value: payload.Metadata.DocType},
			"tags":         &types.AttributeValueMemberS{Value: string(tags)},
			"source":       &types.AttributeValueMemberS{Value: payload.Metadata.Source},
			"visibility":   &types.AttributeValueMemberS{Value: payload.Metadata.Visibility},
			"model":        &types.AttributeValueMemberS{Value: payload.Embedding.Model},
			"strategy":     &types.AttributeValueMemberS{Value: payload.Chunking.Method},
			"chunk_count":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", len(chunks))},
			"deployed_at":  &types.AttributeValueMemberS{Value: payload.DeployedAt},
			"created_at":   &types.AttributeValueMemberS{Value: now},
			"execution_id": &types.AttributeValueMemberS{Value: input.ExecutionID},
		},
	})

	// 4. Deleta temp
	updateStatus(ctx, input.ExecutionID, "save_metadata", "running", 98, "Limpando arquivos temporários...")
	for _, key := range []string{
		fmt.Sprintf("temp/%s/text.txt", input.ExecutionID),
		fmt.Sprintf("temp/%s/chunks.json", input.ExecutionID),
		fmt.Sprintf("temp/%s/vectors.json", input.ExecutionID),
	} {
		s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	}

	updateStatus(ctx, input.ExecutionID, "save_metadata", "done", 100,
		fmt.Sprintf("Pipeline completo — %d chunks salvos", len(chunks)))
	log.Printf("[save-metadata] done — document_id=%s", documentID)

	return map[string]interface{}{
		"document_id": documentID,
		"chunk_count": len(chunks),
		"status":      "done",
	}, nil
}

// putVectorsREST chama a API REST do S3 Vectors com AWS Signature V4
func putVectorsREST(ctx context.Context, items []S3VectorItem) error {
	body := PutVectorsBody{
		VectorBucketName: vectorBucket,
		IndexName:        vectorIndex,
		Vectors:          items,
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	url := fmt.Sprintf("https://s3vectors.%s.api.aws/PutVectors", awsRegion)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyJSON))
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	creds, err := awsCfg.Credentials.Retrieve(ctx)
	if err != nil {
		return fmt.Errorf("credentials error: %w", err)
	}

	if err := signer.SignHTTP(ctx, creds, req, hashBody(bodyJSON), "s3vectors", awsRegion, time.Now()); err != nil {
		return fmt.Errorf("sign error: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		var errBody bytes.Buffer
		errBody.ReadFrom(resp.Body)
		return fmt.Errorf("s3vectors status %d: %s", resp.StatusCode, errBody.String())
	}

	return nil
}

func indexForModel(model string) string {
	switch model {
	case "nomic-embed-text":
		return "global-chunks-768"
	case "all-minilm":
		return "global-chunks-384"
	case "amazon.titan-embed-text-v2":
		return "global-chunks-1024"
	default:
		return "global-chunks-768"
	}
}

func hashBody(body []byte) string {
	h := sha256.Sum256(body)
	return hex.EncodeToString(h[:])
}

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

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	lambda.Start(handler)
}

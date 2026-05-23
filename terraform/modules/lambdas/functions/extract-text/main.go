package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ledongthuc/pdf"
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

	log.Printf("[extract-text] init OK — bucket=%s table=%s", bucket, execTable)
}

// ── Tipos ─────────────────────────────────────────────────────────────────────

type Input struct {
	S3Key       string          `json:"s3_key"`
	Bucket      string          `json:"bucket"`
	ExecutionID string          `json:"execution_id"`
	Payload     json.RawMessage `json:"payload"`
}

type Output struct {
	S3Key       string          `json:"s3_key"`
	Bucket      string          `json:"bucket"`
	TempTextKey string          `json:"temp_text_key"`
	ExecutionID string          `json:"execution_id"`
	Payload     json.RawMessage `json:"payload"`
}

// ── Handler ───────────────────────────────────────────────────────────────────

func handler(ctx context.Context, input Input) (Output, error) {
	log.Printf("[extract-text] start — s3_key=%s execution_id=%s", input.S3Key, input.ExecutionID)
	updateStatus(ctx, input.ExecutionID, "extract_text", "running", 10, "Baixando PDF do S3...")

	// Baixa o PDF do S3
	log.Printf("[extract-text] downloading PDF from s3://%s/%s", input.Bucket, input.S3Key)
	result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(input.Bucket),
		Key:    aws.String(input.S3Key),
	})
	if err != nil {
		log.Printf("[extract-text] failed to download PDF: %v", err)
		updateStatus(ctx, input.ExecutionID, "extract_text", "error", 10, err.Error())
		return Output{}, fmt.Errorf("failed to download PDF: %w", err)
	}
	defer result.Body.Close()

	// Salva em /tmp para processar
	tmpPDF := "/tmp/input.pdf"
	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)
	if err := os.WriteFile(tmpPDF, buf.Bytes(), 0644); err != nil {
		log.Printf("[extract-text] failed to write PDF to /tmp: %v", err)
		updateStatus(ctx, input.ExecutionID, "extract_text", "error", 10, err.Error())
		return Output{}, fmt.Errorf("failed to save PDF: %w", err)
	}

	log.Printf("[extract-text] PDF saved to /tmp (%d bytes)", buf.Len())
	updateStatus(ctx, input.ExecutionID, "extract_text", "running", 15, "Extraindo texto do PDF...")

	// Extrai texto com ledongthuc/pdf (pura Go, sem binários externos)
	text, err := extractText(tmpPDF)
	if err != nil {
		log.Printf("[extract-text] failed to extract text: %v", err)
		updateStatus(ctx, input.ExecutionID, "extract_text", "error", 15, err.Error())
		return Output{}, fmt.Errorf("failed to extract text: %w", err)
	}

	log.Printf("[extract-text] text extracted — %d chars", len(text))
	updateStatus(ctx, input.ExecutionID, "extract_text", "running", 20, "Salvando texto no S3...")

	// Salva texto no S3 temp
	tempTextKey := fmt.Sprintf("temp/%s/text.txt", input.ExecutionID)
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(tempTextKey),
		Body:        strings.NewReader(text),
		ContentType: aws.String("text/plain"),
	})
	if err != nil {
		log.Printf("[extract-text] failed to upload text: %v", err)
		updateStatus(ctx, input.ExecutionID, "extract_text", "error", 20, err.Error())
		return Output{}, fmt.Errorf("failed to upload text: %w", err)
	}

	log.Printf("[extract-text] text saved to s3://%s/%s", bucket, tempTextKey)
	updateStatus(ctx, input.ExecutionID, "extract_text", "done", 25,
		fmt.Sprintf("Texto extraído: %d caracteres", len(text)))

	return Output{
		S3Key:       input.S3Key,
		Bucket:      input.Bucket,
		TempTextKey: tempTextKey,
		ExecutionID: input.ExecutionID,
		Payload:     input.Payload,
	}, nil
}

// ── PDF extraction ────────────────────────────────────────────────────────────

func extractText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	var sb strings.Builder
	totalPages := r.NumPage()
	log.Printf("[extract-text] PDF has %d pages", totalPages)

	for i := 1; i <= totalPages; i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			log.Printf("[extract-text] warning: failed to extract page %d: %v", i, err)
			continue
		}
		sb.WriteString(text)
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// ── DynamoDB status ───────────────────────────────────────────────────────────

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
		log.Printf("[extract-text] failed to update DynamoDB: %v", err)
	}
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

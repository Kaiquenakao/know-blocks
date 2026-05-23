package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

var (
	s3Client  *s3.Client
	sfnClient *sfn.Client
	bucket    string
	smARN     string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}
	s3Client = s3.NewFromConfig(cfg)
	sfnClient = sfn.NewFromConfig(cfg)
	bucket = mustEnv("DOCUMENTS_BUCKET")
	smARN = mustEnv("STATE_MACHINE_ARN")

	log.Printf("[pipe-trigger] init OK — bucket=%s smARN=%s", bucket, smARN)
}

// ── Tipos ─────────────────────────────────────────────────────────────────────

type ConfirmBody struct {
	UploadID string          `json:"upload_id"`
	S3Key    string          `json:"s3_key"`
	Payload  json.RawMessage `json:"payload"`
}

type StepFunctionInput struct {
	S3Key       string          `json:"s3_key"`
	Bucket      string          `json:"bucket"`
	ExecutionID string          `json:"execution_id"`
	Payload     json.RawMessage `json:"payload"`
}

// ── Handler ───────────────────────────────────────────────────────────────────

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("[pipe-trigger] method=%s body=%s", req.HTTPMethod, req.Body[:min(len(req.Body), 200)])

	if req.HTTPMethod != http.MethodPost {
		log.Printf("[pipe-trigger] method not allowed: %s", req.HTTPMethod)
		return respond(405, map[string]string{"error": "method_not_allowed"})
	}

	var body ConfirmBody
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		log.Printf("[pipe-trigger] failed to parse body: %v", err)
		return respond(400, map[string]string{
			"error":   "invalid_body",
			"message": err.Error(),
		})
	}

	log.Printf("[pipe-trigger] upload_id=%s s3_key=%s", body.UploadID, body.S3Key)

	if body.S3Key == "" {
		log.Printf("[pipe-trigger] s3_key is empty")
		return respond(400, map[string]string{
			"error":   "missing_s3_key",
			"message": "s3_key é obrigatório",
		})
	}

	// HeadObject com retry — S3 pode demorar alguns segundos para propagar
	var headErr error
	for i := 1; i <= 3; i++ {
		log.Printf("[pipe-trigger] HeadObject attempt %d/3 — bucket=%s key=%s", i, bucket, body.S3Key)
		_, headErr = s3Client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(body.S3Key),
		})
		if headErr == nil {
			log.Printf("[pipe-trigger] HeadObject OK on attempt %d", i)
			break
		}
		log.Printf("[pipe-trigger] HeadObject failed attempt %d: %v", i, headErr)
		time.Sleep(2 * time.Second)
	}

	if headErr != nil {
		log.Printf("[pipe-trigger] file not found after 3 attempts: %v", headErr)
		return respond(400, map[string]string{
			"error":   "file_not_found",
			"message": fmt.Sprintf("Arquivo não encontrado no S3 após 3 tentativas: %v", headErr),
		})
	}

	// Monta input da Step Function
	executionID := body.UploadID // usa upload_id como execution_id
	sfInput := StepFunctionInput{
		S3Key:       body.S3Key,
		Bucket:      bucket,
		ExecutionID: executionID,
		Payload:     body.Payload,
	}
	sfInputJSON, err := json.Marshal(sfInput)
	if err != nil {
		log.Printf("[pipe-trigger] failed to marshal sf input: %v", err)
		return respond(500, map[string]string{"error": err.Error()})
	}

	log.Printf("[pipe-trigger] starting step function — arn=%s input=%s", smARN, string(sfInputJSON)[:min(len(string(sfInputJSON)), 200)])

	// Inicia Step Function
	execution, err := sfnClient.StartExecution(ctx, &sfn.StartExecutionInput{
		StateMachineArn: aws.String(smARN),
		Input:           aws.String(string(sfInputJSON)),
	})
	if err != nil {
		log.Printf("[pipe-trigger] StartExecution failed: %v", err)
		return respond(500, map[string]string{
			"error":   "step_function_failed",
			"message": err.Error(),
		})
	}

	loc, _ := time.LoadLocation("America/Sao_Paulo")
	now := time.Now().In(loc).Format("02/01/2006 15:04:05")

	log.Printf("[pipe-trigger] execution started — arn=%s deployed_at=%s", *execution.ExecutionArn, now)

	return respond(200, map[string]interface{}{
		"execution_arn": *execution.ExecutionArn,
		"s3_key":        body.S3Key,
		"deployed_at":   now,
	})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func respond(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
	b, _ := json.Marshal(body)
	log.Printf("[pipe-trigger] responding %d — %s", statusCode, string(b)[:min(len(string(b)), 200)])
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(b),
	}, nil
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("missing required env var: %s", key))
	}
	return v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	lambda.Start(handler)
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	s3Client  = s3.NewFromConfig(cfg)
	sfnClient = sfn.NewFromConfig(cfg)
	bucket    = mustEnv("DOCUMENTS_BUCKET")
	smARN     = mustEnv("STATE_MACHINE_ARN")
}

// ── Tipos ─────────────────────────────────────────────────────────────────────

type ConfirmBody struct {
	UploadID string          `json:"upload_id"`
	S3Key    string          `json:"s3_key"`
	Payload  json.RawMessage `json:"payload"`
}

type StepFunctionInput struct {
	S3Key   string          `json:"s3_key"`
	Bucket  string          `json:"bucket"`
	Payload json.RawMessage `json:"payload"`
}

// ── Handler ───────────────────────────────────────────────────────────────────

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != http.MethodPost {
		return respond(405, map[string]string{"error": "method_not_allowed"})
	}

	var body ConfirmBody
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		return respond(400, map[string]string{
			"error":   "invalid_body",
			"message": err.Error(),
		})
	}

	if body.S3Key == "" {
		return respond(400, map[string]string{
			"error":   "missing_s3_key",
			"message": "s3_key é obrigatório",
		})
	}

	// Confirma que o arquivo chegou no S3
	_, err := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(body.S3Key),
	})
	if err != nil {
		return respond(400, map[string]string{
			"error":   "file_not_found",
			"message": "Arquivo não encontrado no S3. Faça o upload antes de confirmar.",
		})
	}

	// Monta input da Step Function
	sfInput := StepFunctionInput{
		S3Key:   body.S3Key,
		Bucket:  bucket,
		Payload: body.Payload,
	}
	sfInputJSON, err := json.Marshal(sfInput)
	if err != nil {
		return respond(500, map[string]string{"error": err.Error()})
	}

	// Inicia Step Function
	execution, err := sfnClient.StartExecution(ctx, &sfn.StartExecutionInput{
		StateMachineArn: aws.String(smARN),
		Input:           aws.String(string(sfInputJSON)),
	})
	if err != nil {
		return respond(500, map[string]string{
			"error":   "step_function_failed",
			"message": err.Error(),
		})
	}

	loc, _ := time.LoadLocation("America/Sao_Paulo")
	now := time.Now().In(loc).Format("02/01/2006 15:04:05")

	return respond(200, map[string]interface{}{
		"execution_arn": *execution.ExecutionArn,
		"s3_key":        body.S3Key,
		"deployed_at":   now,
	})
}

func respond(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
	b, _ := json.Marshal(body)
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

func main() {
	lambda.Start(handler)
}

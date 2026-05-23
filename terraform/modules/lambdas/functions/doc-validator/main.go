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
	"github.com/google/uuid"
)

var (
	s3Client      *s3.Client
	presignClient *s3.PresignClient
	bucket        string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}
	s3Client = s3.NewFromConfig(cfg)
	presignClient = s3.NewPresignClient(s3Client)
	bucket = mustEnv("DOCUMENTS_BUCKET")
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != http.MethodGet {
		return respond(405, map[string]string{"error": "method_not_allowed"})
	}

	filename := req.QueryStringParameters["filename"]
	if filename == "" {
		return respond(400, map[string]string{
			"error":   "missing_param",
			"message": "filename é obrigatório",
		})
	}

	s3Key := fmt.Sprintf("documents/%s", filename)

	// Verifica se já existe no S3
	_, err := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(s3Key),
	})
	if err == nil {
		return respond(409, map[string]string{
			"error":   "document_already_exists",
			"message": fmt.Sprintf("%s já foi deployado anteriormente.", filename),
			"s3_key":  s3Key,
		})
	}

	// Gera presigned URL válida por 15 min
	presigned, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(s3Key),
		ContentType: aws.String("application/pdf"),
	}, func(o *s3.PresignOptions) {
		o.Expires = 15 * time.Minute
	})
	if err != nil {
		return respond(500, map[string]string{
			"error":   "presign_failed",
			"message": err.Error(),
		})
	}

	return respond(200, map[string]string{
		"upload_id":     uuid.New().String(),
		"presigned_url": presigned.URL,
		"s3_key":        s3Key,
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

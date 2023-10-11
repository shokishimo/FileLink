package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/google/uuid"
)

type APIHandler struct {
	AwsConfig 			aws.Config
	dbClient  			*dynamodb.Client
	s3Client  			*s3.Client
	s3PresignClient *s3.PresignClient
}

const (
	awsRegion       string = "us-east-2"
	dynamoTableName string = "FileLinkDB"
	s3BucketName    string = "file-link-s3bucket"
	lifetimeSecs    int64    = 180
)

type PresignedStruct struct {
	Url string `json:"url"`
}

var apiHandler APIHandler

func main() {
	awsConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	if err != nil {
		panic(err)
	}
	apiHandler = APIHandler{
		AwsConfig: awsConfig,
		dbClient:  dynamodb.NewFromConfig(awsConfig),
		s3Client: s3.NewFromConfig(awsConfig),
		s3PresignClient: s3.NewPresignClient(s3.NewFromConfig(awsConfig)),
	}

	http.HandleFunc("/api/generatePresignedUrl", generatePresignedUrl)

	lambda.Start(httpadapter.New(http.DefaultServeMux).ProxyWithContext)
}

func generatePresignedUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}

	// Create the Presigned URL
	objectKey := uuid.New().String()
	request, err := apiHandler.s3PresignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(objectKey),
	}, func(opt *s3.PresignOptions) {
		opt.Expires = time.Duration(lifetimeSecs * int64(time.Second))
	})
	if err != nil {
		fmt.Println(fmt.Printf("Couldn't get a presigned request to get %v:%v. Here's why: %v\n", s3BucketName, objectKey, err))
		http.Error(w, fmt.Sprintf("Failed to generate presigned URL: %v", err), http.StatusInternalServerError)
		return
	}

	response := PresignedStruct{
		Url: request.URL,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Println(fmt.Printf("Couldn't Marshal a struct: %s", response.Url))
		http.Error(w, fmt.Sprintf("Failed to generate presigned URL: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBytes)
}
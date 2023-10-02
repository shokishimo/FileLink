package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"encoding/json"

	// "bytes"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/google/uuid"
)

type APIHandler struct {
	AwsConfig aws.Config
	dbClient  *dynamodb.Client
	s3Client  *s3.Client
}

const (
	awsRegion       string = "us-east-2"
	dynamoTableName string = "FileLinkDB"
	s3BucketName    string = "file-link-s3bucket"
)

type Url struct {
	UrlKey string `json:"url_key"`
}

func main() {
	awsConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	if err != nil {
		panic(err)
	}
	apiHandler := APIHandler{
		AwsConfig: awsConfig,
		dbClient: dynamodb.NewFromConfig(awsConfig),
		s3Client: s3.NewFromConfig(awsConfig),
	}

	http.HandleFunc("/createNewUrl", func(w http.ResponseWriter, r *http.Request) {
		if (r.Method != http.MethodGet) {
			http.Error(w, "Method is not alloed", http.StatusMethodNotAllowed)
			return
		}
		u := uuid.New()

		response := Url{
			UrlKey: u.String(),
		}

		// Serialize the response object to JSON
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
			return
		}
		
		// Set Content-Type and send the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	})

	http.HandleFunc("/share/", func(w http.ResponseWriter, r *http.Request) {
		if (r.Method != http.MethodPost) {
			http.Error(w, "Method is not alloed", http.StatusMethodNotAllowed)
			return
		}
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 {
				http.Error(w, "Not found", http.StatusNotFound)
				return
		}
		urlKey := pathParts[2]

		// Retrieve the file from post body
		file, _, err := r.FormFile("zip-file")
		if err != nil {
			http.Error(w, "Unable to get the file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Upload the file to S3
		_, err = apiHandler.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
			Bucket: 		 aws.String(s3BucketName),
			Key:    		 aws.String("Filename_" + urlKey),
			Body:   		 file,
			ContentType: aws.String("application/zip"),
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to upload to S3: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		// success
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Successfully uploaded file to S3"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if (r.Method == http.MethodGet) {
			io.WriteString(w, "root with Get")
			return
		}
		if (r.Method == http.MethodPost) {
			io.WriteString(w, "root with Post")
			return
		}
	})
	
	lambda.Start(httpadapter.New(http.DefaultServeMux).ProxyWithContext)
}
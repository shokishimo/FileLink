package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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
		dbClient:  dynamodb.NewFromConfig(awsConfig),
		s3Client:  s3.NewFromConfig(awsConfig),
	}

// GET
	http.HandleFunc("/api/createNewUrl", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method is not alloed", http.StatusMethodNotAllowed)
			return
		}

		u := uuid.New()
		resUrl := Url{
			UrlKey: u.String(),
		}
		jsonRes, err := json.Marshal(resUrl)
		if err != nil {
			http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonRes)
	})

// POST
	http.HandleFunc("/api/share/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method is not alloed", http.StatusMethodNotAllowed)
			return
		}
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		urlKey := pathParts[2]

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			fmt.Printf("Error parsing multipart form: %s\n", err.Error())
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		for key, fileHeaders := range r.MultipartForm.File {
			for _, fileHeader := range fileHeaders {
				file, err := fileHeader.Open()
				if err != nil {
					fmt.Printf("Error reading file: %s\n", err.Error())
					http.Error(w, "Error reading file", http.StatusInternalServerError)
					return
				}
	
				if err := uploadToS3(apiHandler.s3Client, s3BucketName, fmt.Sprintf("%s_%s", urlKey, key), file); err != nil {
					file.Close()
					fmt.Printf("Error uploading to S3: %s\n", err.Error())
					http.Error(w, "Error uploading to S3", http.StatusInternalServerError)
					return
				}		
				defer file.Close()
			}
		}

		// success
		resArray := []string{"success ", "success"} // TODO: convert these to actual url strings
		jsonRes, err := json.Marshal(resArray)
		if err != nil {
				http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
				return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonRes)
	})

	lambda.Start(httpadapter.New(http.DefaultServeMux).ProxyWithContext)
}


func uploadToS3(s3Client *s3.Client, bucketName string, key string, body io.Reader) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   body,
		ContentType: aws.String("application/zip"),
	}

	_, err := s3Client.PutObject(context.TODO(), input)
	return err
}
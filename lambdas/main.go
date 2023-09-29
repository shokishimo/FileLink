package main

import (
	"context"
	"log"
	"fmt"
	"net/http"
	"io"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

type APIHandler struct {
	AwsConfig aws.Config
	DynamoTableName string
	dbClient  *dynamodb.Client
	S3BucketName string
	s3Client *s3.Client
}


func main() {
	log.Printf("Fiber cold start")
	awsConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-2"))
	if err != nil {
		panic(err)
	}
	apiHandler := APIHandler{
		AwsConfig: awsConfig,
		DynamoTableName: "FileLinkDB",
		dbClient: dynamodb.NewFromConfig(awsConfig),
		S3BucketName: "file-link-s3bucket",
		s3Client: s3.NewFromConfig(awsConfig),
	}

	http.HandleFunc("/share/", func(w http.ResponseWriter, r *http.Request) {
		if (r.Method == http.MethodPost) {
			pathParts := strings.Split(r.URL.Path, "/")
			if len(pathParts) < 3 {
					http.Error(w, "Not found", http.StatusNotFound)
					return
			}
			valueAfterShare := pathParts[2]
			fmt.Fprintf(w, "Value after /share/ is: %s", valueAfterShare)

			// Parse the form data to retrieve the file
			err := r.ParseMultipartForm(10 << 20) // 10 MB limit
			if err != nil {
				http.Error(w, "Unable to parse form", http.StatusBadRequest)
				return
			}

			// Retrieve the file from post body
			file, fileHeader, err := r.FormFile("zip-file")
			if err != nil {
				http.Error(w, "Unable to get the file", http.StatusBadRequest)
				return
			}
			defer file.Close()

			// Upload the file to S3
			res, err := apiHandler.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
				Bucket: &apiHandler.S3BucketName,
				Key:    aws.String(fileHeader.Filename),
				Body:   file,
				ContentType: aws.String("application/zip"),
			})
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to upload to S3: %s", err.Error()), http.StatusInternalServerError)
				return
			}
	
			// success
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(fmt.Sprintf("Successfully uploaded file to S3: %v", res)))
			return
		}
	})

	http.HandleFunc("/abc", func(w http.ResponseWriter, r *http.Request) {
		if (r.Method == http.MethodGet) {
			io.WriteString(w, "/abc with Get")
			return
		}
		if (r.Method == http.MethodPost) {
			io.WriteString(w, "/abc with Post")
			return
		}
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
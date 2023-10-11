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
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
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

// GET
	http.HandleFunc("/api/download/", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
			http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
			return
		}

		pathParts := strings.Split(r.URL.Path, "/")
    if len(pathParts) < 4 {
        http.Error(w, "Not found", http.StatusNotFound)
        return
    }
    s3ObjectKey := pathParts[3]
		fmt.Println("Parsed S3 Object Key: ", s3ObjectKey)

		// Fetch the file from S3
		resp, err := downloadFromS3(apiHandler.s3Client, s3BucketName, s3ObjectKey)
		if err != nil {
				fmt.Printf("Error downloading from S3: %s\n", err.Error())
				http.Error(w, "Error downloading from S3", http.StatusInternalServerError)
				return
		}
    defer resp.Body.Close()

    // Set the appropriate headers
    w.Header().Set("Content-Type", aws.ToString(resp.ContentType))
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; ObjectKey=%s", s3ObjectKey))

    // Stream the file to the HTTP response
    io.Copy(w, resp.Body)
	})

// POST
	http.HandleFunc("/api/share/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			fmt.Printf("Method is not alloed: %s\n", err.Error())
			http.Error(w, "Method is not alloed", http.StatusMethodNotAllowed)
			return
		}
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 4 {
			fmt.Printf("Not found: %s\n", err.Error())
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		urlKey := pathParts[3]

		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			fmt.Printf("Couldn't parse multiform: %s\n", err.Error())
			http.Error(w, "Couldn't parse multiform", http.StatusNotFound)
			return
		}

		// Retrieve the file from post body
		file, fileHeader, err := r.FormFile("theFile")
		if err != nil {
			fmt.Printf("Unable to get the file3: %s\n", err.Error())
			http.Error(w, "Unable to get the file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		ind := "0"
		res, err := uploadToS3(apiHandler.s3Client, s3BucketName, fmt.Sprintf("%s_%s", urlKey, ind), file, fileHeader.Header.Get("Content-Type"))
		if err != nil {
			fmt.Printf("Error uploading to S3: %s\n", err.Error())
			http.Error(w, "Error uploading to S3", http.StatusInternalServerError)
			return
		}

		// Construct the S3 URL for the uploaded file
		var uploadedUrls []string
		uploadedUrls = append(uploadedUrls, "https://file-link-s3bucket.s3.us-east-2.amazonaws.com/" + *res.Key)

		// success
		jsonRes, err := json.Marshal(uploadedUrls)
		if err != nil {
			fmt.Printf("Failed to serialize response: %s\n", err.Error())
			http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonRes)
	})

	lambda.Start(httpadapter.New(http.DefaultServeMux).ProxyWithContext)
}



func downloadFromS3(s3Client *s3.Client, bucketName string, key string) (*s3.GetObjectOutput, error) {
	input := &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
	}

	resp, err := s3Client.GetObject(context.TODO(), input)
	if err != nil {
			return nil, err
	}

	return resp, nil
}


func uploadToS3(s3Client *s3.Client, bucketName string, key string, body io.Reader, contentType string) (*manager.UploadOutput, error) {
	uploader := manager.NewUploader(s3Client)
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   body,
		ContentType: aws.String(contentType),
	}

	res, err := uploader.Upload(context.TODO(), input)
	return res, err
}
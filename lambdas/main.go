package main

import (
	//"context"
	"log"
	//"fmt"
	"net/http"
	"io"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	//"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	//"github.com/gofiber/fiber/v2"
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
	// awsConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-2"))
	// if err != nil {
	// 	panic(err)
	// }
	// apiHandler := APIHandler{
	// 	AwsConfig: awsConfig,
	// 	DynamoTableName: "FileLinkDB",
	// 	dbClient: dynamodb.NewFromConfig(awsConfig),
	// 	S3BucketName: "file-link-bucket",
	// 	s3Client: s3.NewFromConfig(awsConfig),
	// }

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

	// app.Post("/share/:path", func(c *fiber.Ctx) error {
	// 	givenFile, err := c.FormFile("zip-file")
	// 	if err != nil {
	// 		return c.Status(500).SendString(err.Error())
	// 	}

	// 	// file type check

	// 	// create a read stream to the uploaded file content to grab the contents
	// 	uploadedFile, err := givenFile.Open()
	// 	if err != nil {
	// 		return c.Status(500).SendString("Could not read file")
	// 	}
	// 	defer uploadedFile.Close()

	// 	// object to upload
	// 	objectInput := &s3.PutObjectInput{
	// 		Bucket: aws.String(apiHandler.S3BucketName),
	// 		Key:    aws.String(givenFile.Filename),
	// 		Body:   uploadedFile,
	// 		ACL:    "public-read",
	// 	}

	// 	// upload to S3
	// 	res, err :=apiHandler.s3Client.PutObject(context.TODO(), objectInput);
	// 	if err != nil {
	// 		return c.Status(500).SendString(fmt.Sprintf("Failed to upload file to S3: %v", err))
	// 	}

	// 	// success
	// 	return c.Status(200).SendString(fmt.Sprintf("Successfully uploaded file to S3: %v", res))
	// })
	
	lambda.Start(httpadapter.New(http.DefaultServeMux).ProxyWithContext)
}
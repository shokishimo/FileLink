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
	//"github.com/aws/aws-sdk-go-v2/service/s3/types"
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
	lifetimeSecs    int64  = 60
)

type PostPresignedRes struct {
	Urls []string `json:"urls"`
	ObjectKeys []string `json:"objectKeys"`
}

type PostPresignedReq struct {
	Num int `json:"numOfFiles"`
}

type GetPresignReq struct {
	Keys []string `json:"keys"`
}

type GetPresignRes struct {
	Urls []string `json:"urls"`
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

	http.HandleFunc("/api/postPresignedUrls", postPresignedUrls)
	http.HandleFunc("/api/getPresignedUrls", getPresignedUrls)
	//http.HandleFunc("/api/emptyBucket", emptyBucket)

	lambda.Start(httpadapter.New(http.DefaultServeMux).ProxyWithContext)
}


// POST
func postPresignedUrls(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}

	var reqBody PostPresignedReq
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %s", err.Error()), http.StatusBadRequest)
		return
	}

	response := PostPresignedRes{
		Urls: make([]string, reqBody.Num),
		ObjectKeys: make([]string, reqBody.Num),
	}

	for i := 0; i < reqBody.Num; i++ {
		// Create the Presigned URLs
		objectKey := uuid.New().String()
		request, err := apiHandler.s3PresignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(s3BucketName),
			Key:    aws.String(objectKey),
		}, func(opt *s3.PresignOptions) {
			opt.Expires = time.Duration(lifetimeSecs * int64(time.Second))
		})
		if err != nil {
			fmt.Println(fmt.Printf("Couldn't get a presigned request (#%v) to get %v:%v. Here's why: %v\n", i, s3BucketName, objectKey, err))
			http.Error(w, fmt.Sprintf("Failed to generate presigned URL: %v", err), http.StatusInternalServerError)
			return
		}

		response.Urls[i] = request.URL
		response.ObjectKeys[i] = objectKey
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Println(fmt.Printf("Couldn't Marshal a struct: %v", response))
		http.Error(w, fmt.Sprintf("Failed to generate presigned URL: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBytes)
}


// POST
func getPresignedUrls(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}

	var reqBody GetPresignReq
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// makes a presigned request that can be used to get an object from a bucket.
	// The presigned request is valid for the specified number of seconds.
	response := new(GetPresignRes)
	for i, key := range reqBody.Keys {
		request, err := apiHandler.s3PresignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: aws.String(s3BucketName),
			Key:    aws.String(key),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = time.Duration(lifetimeSecs * 60 * int64(time.Second))
		})
		if err != nil {
			fmt.Println(fmt.Printf("Couldn't get a presigned request (#%v) to get %v:%v. Here's why: %v\n", i, s3BucketName, key, err))
			http.Error(w, fmt.Sprintf("Failed to generate presigned URL: %v", err), http.StatusInternalServerError)
			return
		}
		response.Urls = append(response.Urls, request.URL)
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Println(fmt.Printf("Couldn't Marshal a struct: %v", response))
		http.Error(w, fmt.Sprintf("Failed to generate presigned URL: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBytes)
}


// DELETE
// func emptyBucket(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodDelete {
// 		w.WriteHeader(http.StatusMethodNotAllowed)
// 		w.Write([]byte("Method not allowed"))
// 		return
// 	}

// 	for {
// 		input := &s3.ListObjectsV2Input{
// 			Bucket:  aws.String(s3BucketName),
// 		}
// 		result, err := apiHandler.s3Client.ListObjectsV2(context.TODO(), input)
// 		if err != nil {
// 			fmt.Println(fmt.Printf("Failed to list objects: %s", err.Error()))
// 			http.Error(w, fmt.Sprintf("Failed to list objects: %s", err.Error()), http.StatusInternalServerError)
// 			return
// 		}
// 		if len(result.Contents) == 0 { // if already empty
// 			break
// 		}

// 		var objectIds []types.ObjectIdentifier
// 		for _, object := range result.Contents {
// 			objectIds = append(objectIds, types.ObjectIdentifier{Key: aws.String(*object.Key)})
// 		}
// 		deleteInput := &s3.DeleteObjectsInput{
// 			Bucket: aws.String(s3BucketName),
// 			Delete: &types.Delete{Objects: objectIds},
// 		}
// 		_, err = apiHandler.s3Client.DeleteObjects(context.TODO(), deleteInput)
// 		if err != nil {
// 			fmt.Println(fmt.Printf("Failed to delete objects: %s", err.Error()))
// 			http.Error(w, fmt.Sprintf("Failed to delete objects: %s", err.Error()), http.StatusInternalServerError)
// 			return
// 		}
// 	}

// 	w.Header().Set("Content-Type", "text/plain")
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte("Successfully empty the bucket"))
// }
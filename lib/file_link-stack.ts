import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as apigateway from "aws-cdk-lib/aws-apigateway";
import * as iam from "aws-cdk-lib/aws-iam";
import * as path from "path";

export class FileLinkStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // S3 Bucket Configuration
    const s3Bucket = new s3.Bucket(this, "S3Bucket", {
      bucketName: "file-link-s3bucket",
      // publicReadAccess: true,
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      enforceSSL: true,
      versioned: true, // Ensures new versions of objects are created on overwrite
      cors: [
        {
          allowedMethods: [s3.HttpMethods.GET, s3.HttpMethods.POST, s3.HttpMethods.PUT, s3.HttpMethods.DELETE],
          allowedOrigins: ["*"],
          allowedHeaders: ["*"],
        }
      ]
    });

    const dynamoTable = new dynamodb.Table(this, "DynamoTable", {
      tableName: "FileLinkDB",
      partitionKey: { name: "ID", type: dynamodb.AttributeType.STRING },
      sortKey: { name: "Path", type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
    });

    dynamoTable.addGlobalSecondaryIndex({
      indexName: "Path-ID-Index",
      partitionKey: { name: "Path", type: dynamodb.AttributeType.STRING },
      sortKey: { name: "ID", type: dynamodb.AttributeType.STRING },
    });

    const lambdaFunc = new lambda.Function(this, "LambdaAPI", {
      runtime: lambda.Runtime.GO_1_X,
      code: lambda.Code.fromAsset(path.join(__dirname, '../lambdas'), {
        assetHashType: cdk.AssetHashType.OUTPUT,
        bundling: {
          image: cdk.DockerImage.fromRegistry('golang:1.19'),
          command: [
            'bash',
            '-c',
            'GOCACHE=/tmp/go-build GOOS=linux GOARCH=amd64 GOFLAGS=-buildvcs=false go build -o /asset-output/main .',
          ],
        },
      }),
      handler: "main",
      memorySize: 3008,
      timeout: cdk.Duration.minutes(15),
      environment: {
        s3Bucket: s3Bucket.bucketName,
        dynamoTable: dynamoTable.tableName,
      }
    });

    s3Bucket.grantReadWrite(lambdaFunc);
    dynamoTable.grantReadWriteData(lambdaFunc);
    
    const api = new apigateway.LambdaRestApi(this, "apiGateway", {
      handler: lambdaFunc,
      proxy: true,
      defaultCorsPreflightOptions: {
        allowOrigins: apigateway.Cors.ALL_ORIGINS,
        allowMethods: apigateway.Cors.ALL_METHODS,
      }
    })
  }
}

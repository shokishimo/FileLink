import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as path from "path";


export class FileLinkStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const s3Bucket = new s3.Bucket(this, "S3Bucket", {
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      removalPolicy: cdk.RemovalPolicy.RETAIN
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
      code: lambda.Code.fromAsset(path.join(__dirname, '../lambdas')),
      handler: "main",
      memorySize: 3008,
      timeout: cdk.Duration.minutes(15),
      environment: {
        s3Bucket: s3Bucket.bucketName,
      }
    });

    s3Bucket.grantReadWrite(lambdaFunc);
    dynamoTable.grantReadWriteData(lambdaFunc);

    lambdaFunc.addFunctionUrl({
      authType: lambda.FunctionUrlAuthType.NONE
    });
  }
}

import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as path from "path";


export class FileLinkStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const s3Bucket = new s3.Bucket(this, "S3Bucket", {
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      removalPolicy: cdk.RemovalPolicy.RETAIN
    });

    const lambdaFunc = new lambda.Function(this, "LambdaAPI", {
      runtime: lambda.Runtime.GO_1_X,
      code: lambda.Code.fromAsset(path.join(__dirname, '../lambdas')),
      handler: "main",
      memorySize: 512,
      timeout: cdk.Duration.seconds(30),
      environment: {
        s3Bucket: s3Bucket.bucketName,
      }
    });

    s3Bucket.grantReadWrite(lambdaFunc);

    lambdaFunc.addFunctionUrl({
      authType: lambda.FunctionUrlAuthType.NONE
    });
  }
}

import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as lambda from "aws-cdk-lib/aws-lambda"
import * as path from "path";


export class FileLinkStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const lambdaFunc = new lambda.Function(this, "LambdaAPI", {
      runtime: lambda.Runtime.GO_1_X,
      code: lambda.Code.fromAsset(path.join(__dirname, '../lambdas')),
      handler: "main",
      memorySize: 512,
      timeout: cdk.Duration.seconds(30),
      environment: {

      }
    });

    lambdaFunc.addFunctionUrl({
      authType: lambda.FunctionUrlAuthType.NONE
    });
  }
}

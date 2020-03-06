{
  build(DeepAlertStackName, SecretArn, LambdaRoleArn, VpcConfig={}):: {
    local ReportTopic = {
      'Fn::ImportValue': DeepAlertStackName + '-ReportTopic',
    },

    AWSTemplateFormatVersion: '2010-09-09',
    Transform: 'AWS::Serverless-2016-10-31',

    Resources: {
      // --------------------------------------------------------
      // Lambda functions
      Handler: {
        Type: 'AWS::Serverless::Function',
        Properties: {
          CodeUri: 'build',
          Handler: 'main',
          Runtime: 'go1.x',
          Timeout: 30,
          MemorySize: 128,
          Role: LambdaRoleArn,
          Environment: {
            Variables: {
              SECRET_ARN: SecretArn,
            },
          },
          Events: {
            NotifyTopic: {
              Type: 'SNS',
              Properties: {
                Topic: ReportTopic,
              },
            },
          },
        } + (if VpcConfig == {} then {} else { VpcConfig: VpcConfig }),
      },
    },
  },
}

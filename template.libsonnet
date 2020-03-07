{
  build(DeepAlertStackName, SecretArn, LambdaRoleArn, VpcConfig={}):: {
    local TaskTopic = { 'Fn::ImportValue': DeepAlertStackName + '-TaskTopic' },
    local ContentQueue = { 'Fn::ImportValue': DeepAlertStackName + '-ContentQueue' },
    local AttributeQueue = { 'Fn::ImportValue': DeepAlertStackName + '-AttributeQueue' },


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
              CONTENT_QUEUE: ContentQueue,
              ATTRIBUTE_QUEUE: AttributeQueue,
            },
          },
          Events: {
            NotifyTopic: {
              Type: 'SNS',
              Properties: { Topic: TaskTopic },
            },
          },
        } + (if VpcConfig == {} then {} else { VpcConfig: VpcConfig }),
      },
    },
  },
}

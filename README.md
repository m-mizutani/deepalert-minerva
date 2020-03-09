# deepalert-minerva

## Setup

### Set Secrets into SecretsManager

Following fields are required.

- `minerva_apikey`
- `minerva_endpoint`
- `strix_endpoint`

### Configuration Files

deploy.jsonnet
```jsonnet
{
  StackName: 'YOUR_STACK_NAME',
  CodeS3Bucket: 'YOUR_BUCKET_NAME',
  CodeS3Prefix: 'functions',
  Region: 'ap-northeast-1',
}
```

stack.jsonnet
```jsonnet
local template = import 'template.libsonnet';

local vpcConfig = {
  SecurityGroupIds: ['sg-b2xxxxx'],
  SubnetIds: ['subnet-abxxxxxx', 'subnet-cdxxxxxx'],
};
local secretArn = 'arn:aws:secretsmanager:ap-northeast-1:12345xxxx:secret:YOUR_SECRETS';

template.build(
  DeepAlertStackName='YOUR_STACK_NAME,
  SecretArn=secretArn,
  VpcConfig=vpcConfig,
  LambdaRoleArn='YOUR_LAMBDA_ROLE_ARN',
)
```

## Deploy

```
$ make -f path/to/Makefile
```

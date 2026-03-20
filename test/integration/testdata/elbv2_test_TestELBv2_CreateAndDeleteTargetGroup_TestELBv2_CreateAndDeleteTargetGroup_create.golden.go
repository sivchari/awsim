{
  "TargetGroups": [
    {
      "HealthCheckEnabled": true,
      "HealthCheckIntervalSeconds": 30,
      "HealthCheckPath": "/",
      "HealthCheckPort": "traffic-port",
      "HealthCheckProtocol": "HTTP",
      "HealthCheckTimeoutSeconds": 5,
      "HealthyThresholdCount": 5,
      "IpAddressType": "",
      "LoadBalancerArns": [],
      "Matcher": null,
      "Port": 80,
      "Protocol": "HTTP",
      "ProtocolVersion": null,
      "TargetControlPort": null,
      "TargetGroupArn": "arn:aws:elasticloadbalancing:us-east-1:000000000000:targetgroup/test-target-group/0b723907-7923-4e4",
      "TargetGroupName": "test-target-group",
      "TargetType": "instance",
      "UnhealthyThresholdCount": 2,
      "VpcId": "vpc-12345678"
    }
  ],
  "ResultMetadata": {}
}
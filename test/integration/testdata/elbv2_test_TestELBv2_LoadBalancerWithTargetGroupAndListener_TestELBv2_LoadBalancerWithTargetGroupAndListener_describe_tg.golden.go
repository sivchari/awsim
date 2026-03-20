{
  "NextMarker": null,
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
      "TargetGroupArn": "arn:aws:elasticloadbalancing:us-east-1:000000000000:targetgroup/test-full-tg/52498d7e-f000-485",
      "TargetGroupName": "test-full-tg",
      "TargetType": "instance",
      "UnhealthyThresholdCount": 2,
      "VpcId": "vpc-12345678"
    }
  ],
  "ResultMetadata": {}
}
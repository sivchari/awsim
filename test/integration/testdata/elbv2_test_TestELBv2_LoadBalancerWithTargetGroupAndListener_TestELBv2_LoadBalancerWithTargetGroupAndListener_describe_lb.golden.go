{
  "LoadBalancers": [
    {
      "AvailabilityZones": [
        {
          "LoadBalancerAddresses": null,
          "OutpostId": null,
          "SourceNatIpv6Prefixes": null,
          "SubnetId": "subnet-12345678",
          "ZoneName": "us-east-1a"
        },
        {
          "LoadBalancerAddresses": null,
          "OutpostId": null,
          "SourceNatIpv6Prefixes": null,
          "SubnetId": "subnet-87654321",
          "ZoneName": "us-east-1b"
        }
      ],
      "CanonicalHostedZoneId": "Z35SXDOTRQ7X7K",
      "CreatedTime": "2026-03-23T07:45:25.754Z",
      "CustomerOwnedIpv4Pool": null,
      "DNSName": "test-full-lb-a916ea63.us-east-1.elb.amazonaws.com",
      "EnablePrefixForIpv6SourceNat": "",
      "EnforceSecurityGroupInboundRulesOnPrivateLinkTraffic": null,
      "IpAddressType": "ipv4",
      "IpamPools": null,
      "LoadBalancerArn": "arn:aws:elasticloadbalancing:us-east-1:000000000000:loadbalancer/app/test-full-lb/a916ea63-b5a6-40c",
      "LoadBalancerName": "test-full-lb",
      "Scheme": "internet-facing",
      "SecurityGroups": [],
      "State": {
        "Code": "active",
        "Reason": null
      },
      "Type": "application",
      "VpcId": "vpc-ac17815b"
    }
  ],
  "NextMarker": null,
  "ResultMetadata": {}
}
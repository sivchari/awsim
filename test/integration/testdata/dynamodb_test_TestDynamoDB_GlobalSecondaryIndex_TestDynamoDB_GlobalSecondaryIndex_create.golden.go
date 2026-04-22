{
  "TableDescription": {
    "ArchivalSummary": null,
    "AttributeDefinitions": [
      {
        "AttributeName": "pk",
        "AttributeType": "S"
      },
      {
        "AttributeName": "sk",
        "AttributeType": "S"
      },
      {
        "AttributeName": "gsi_pk",
        "AttributeType": "S"
      }
    ],
    "BillingModeSummary": {
      "BillingMode": "PAY_PER_REQUEST",
      "LastUpdateToPayPerRequestDateTime": null
    },
    "CreationDateTime": "2026-04-22T01:08:47Z",
    "DeletionProtectionEnabled": false,
    "GlobalSecondaryIndexes": [
      {
        "Backfilling": null,
        "IndexArn": "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-gsi/index/gsi-index",
        "IndexName": "gsi-index",
        "IndexSizeBytes": 0,
        "IndexStatus": "ACTIVE",
        "ItemCount": 0,
        "KeySchema": [
          {
            "AttributeName": "gsi_pk",
            "KeyType": "HASH"
          }
        ],
        "OnDemandThroughput": null,
        "Projection": {
          "NonKeyAttributes": null,
          "ProjectionType": "ALL"
        },
        "ProvisionedThroughput": null,
        "WarmThroughput": null
      }
    ],
    "GlobalTableSettingsReplicationMode": "",
    "GlobalTableVersion": null,
    "GlobalTableWitnesses": null,
    "ItemCount": 0,
    "KeySchema": [
      {
        "AttributeName": "pk",
        "KeyType": "HASH"
      },
      {
        "AttributeName": "sk",
        "KeyType": "RANGE"
      }
    ],
    "LatestStreamArn": null,
    "LatestStreamLabel": null,
    "LocalSecondaryIndexes": null,
    "MultiRegionConsistency": "",
    "OnDemandThroughput": null,
    "ProvisionedThroughput": null,
    "Replicas": null,
    "RestoreSummary": null,
    "SSEDescription": null,
    "StreamSpecification": null,
    "TableArn": "arn:aws:dynamodb:us-east-1:000000000000:table/test-table-gsi",
    "TableClassSummary": null,
    "TableId": "ad896e2b-c27f-45be-a28a-8f444a5560d9",
    "TableName": "test-table-gsi",
    "TableSizeBytes": 0,
    "TableStatus": "ACTIVE",
    "WarmThroughput": null
  },
  "ResultMetadata": {}
}
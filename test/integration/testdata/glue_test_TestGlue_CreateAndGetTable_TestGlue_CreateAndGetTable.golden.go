{
  "Table": {
    "Name": "test_table",
    "CatalogId": null,
    "CreateTime": "2026-03-23T07:45:25.974Z",
    "CreatedBy": null,
    "DatabaseName": "table_test_database",
    "Description": "Test table",
    "FederatedTable": null,
    "IsMaterializedView": null,
    "IsMultiDialectView": null,
    "IsRegisteredWithLakeFormation": false,
    "LastAccessTime": null,
    "LastAnalyzedTime": null,
    "Owner": null,
    "Parameters": null,
    "PartitionKeys": null,
    "Retention": 0,
    "Status": null,
    "StorageDescriptor": {
      "AdditionalLocations": null,
      "BucketColumns": null,
      "Columns": [
        {
          "Name": "id",
          "Comment": null,
          "Parameters": null,
          "Type": "int"
        },
        {
          "Name": "name",
          "Comment": null,
          "Parameters": null,
          "Type": "string"
        }
      ],
      "Compressed": false,
      "InputFormat": "org.apache.hadoop.mapred.TextInputFormat",
      "Location": "s3://test-bucket/data/",
      "NumberOfBuckets": 0,
      "OutputFormat": "org.apache.hadoop.hive.ql.io.HiveIgnoreKeyTextOutputFormat",
      "Parameters": null,
      "SchemaReference": null,
      "SerdeInfo": null,
      "SkewedInfo": null,
      "SortColumns": null,
      "StoredAsSubDirectories": false
    },
    "TableType": "EXTERNAL_TABLE",
    "TargetTable": null,
    "UpdateTime": "2026-03-23T07:45:25.974Z",
    "VersionId": null,
    "ViewDefinition": null,
    "ViewExpandedText": null,
    "ViewOriginalText": null
  },
  "ResultMetadata": {}
}
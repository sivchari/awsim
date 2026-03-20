{
  "CostCategory": {
    "CostCategoryArn": "arn:aws:ce::123456789012:costcategory/d9b7d908-3a4c-4085-a8fb-bf1d59a3ccb2",
    "EffectiveStart": "2026-03-23T07:45:25Z",
    "Name": "test-cost-category",
    "RuleVersion": "CostCategoryExpression.v1",
    "Rules": [
      {
        "InheritedValue": null,
        "Rule": {
          "And": null,
          "CostCategories": null,
          "Dimensions": {
            "Key": "LINKED_ACCOUNT",
            "MatchOptions": null,
            "Values": [
              "123456789012"
            ]
          },
          "Not": null,
          "Or": null,
          "Tags": null
        },
        "Type": "",
        "Value": "Development"
      }
    ],
    "DefaultValue": "Other",
    "EffectiveEnd": null,
    "ProcessingStatus": [
      {
        "Component": "COST_EXPLORER",
        "Status": "APPLIED"
      }
    ],
    "SplitChargeRules": null
  },
  "ResultMetadata": {}
}
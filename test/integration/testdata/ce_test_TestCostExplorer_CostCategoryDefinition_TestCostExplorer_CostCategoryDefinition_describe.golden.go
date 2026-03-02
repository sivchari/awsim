{
  "CostCategory": {
    "CostCategoryArn": "arn:aws:ce::123456789012:costcategory/ef848357-b551-43c4-978e-019642cba0ec",
    "EffectiveStart": "2026-03-02T02:06:47Z",
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
{
  "Certificate": {
    "CertificateArn": "arn:aws:acm:us-east-1:000000000000:certificate/17106d9f-c437-4a6c-befe-193d0b61c8fb",
    "CertificateAuthorityArn": null,
    "CreatedAt": "2026-03-23T07:45:24.814Z",
    "DomainName": "options-test.example.com",
    "DomainValidationOptions": [
      {
        "DomainName": "options-test.example.com",
        "HttpRedirect": null,
        "ResourceRecord": {
          "Name": "_acme-challenge.options-test.example.com",
          "Type": "CNAME",
          "Value": "_17106d9f.acm-validations.aws."
        },
        "ValidationDomain": "options-test.example.com",
        "ValidationEmails": null,
        "ValidationMethod": "DNS",
        "ValidationStatus": "PENDING_VALIDATION"
      }
    ],
    "ExtendedKeyUsages": null,
    "FailureReason": "",
    "ImportedAt": null,
    "InUseBy": null,
    "IssuedAt": null,
    "Issuer": null,
    "KeyAlgorithm": "EC_prime256v1",
    "KeyUsages": null,
    "ManagedBy": "",
    "NotAfter": null,
    "NotBefore": null,
    "Options": {
      "CertificateTransparencyLoggingPreference": "ENABLED",
      "Export": ""
    },
    "RenewalEligibility": "INELIGIBLE",
    "RenewalSummary": null,
    "RevocationReason": "",
    "RevokedAt": null,
    "Serial": "d4f59677b399c61fa428759b27299d6b",
    "SignatureAlgorithm": null,
    "Status": "PENDING_VALIDATION",
    "Subject": "CN=options-test.example.com",
    "SubjectAlternativeNames": null,
    "Type": "AMAZON_ISSUED"
  },
  "ResultMetadata": {}
}
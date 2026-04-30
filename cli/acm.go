package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	acmTypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/spf13/cobra"
)

func newACMCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "acm",
		Short: "ACM commands",
	}

	cmd.AddCommand(
		newACMRequestCertificateCmd(),
		newACMDescribeCertificateCmd(),
		newACMListCertificatesCmd(),
		newACMDeleteCertificateCmd(),
		newACMGetCertificateCmd(),
		newACMImportCertificateCmd(),
	)

	return cmd
}

func newACMRequestCertificateCmd() *cobra.Command {
	var domainName, validationMethod, keyAlgorithm string

	cmd := &cobra.Command{
		Use:   "request-certificate",
		Short: "Request an ACM certificate",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := acm.NewFromConfig(cfg, func(o *acm.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &acm.RequestCertificateInput{
				DomainName: aws.String(domainName),
			}

			if validationMethod != "" {
				input.ValidationMethod = acmTypes.ValidationMethod(validationMethod)
			}

			if keyAlgorithm != "" {
				input.KeyAlgorithm = acmTypes.KeyAlgorithm(keyAlgorithm)
			}

			out, err := client.RequestCertificate(cmd.Context(), input)
			if err != nil {
				return fmt.Errorf("request-certificate failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&domainName, "domain-name", "", "Domain name")
	cmd.Flags().StringVar(&validationMethod, "validation-method", "", "Validation method (DNS or EMAIL)")
	cmd.Flags().StringVar(&keyAlgorithm, "key-algorithm", "", "Key algorithm")

	return cmd
}

func newACMDescribeCertificateCmd() *cobra.Command {
	var certificateArn string

	cmd := &cobra.Command{
		Use:   "describe-certificate",
		Short: "Describe an ACM certificate",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := acm.NewFromConfig(cfg, func(o *acm.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.DescribeCertificate(cmd.Context(), &acm.DescribeCertificateInput{
				CertificateArn: aws.String(certificateArn),
			})
			if err != nil {
				return fmt.Errorf("describe-certificate failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&certificateArn, "certificate-arn", "", "Certificate ARN")

	return cmd
}

func newACMListCertificatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-certificates",
		Short: "List ACM certificates",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := acm.NewFromConfig(cfg, func(o *acm.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.ListCertificates(cmd.Context(), &acm.ListCertificatesInput{})
			if err != nil {
				return fmt.Errorf("list-certificates failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func newACMDeleteCertificateCmd() *cobra.Command {
	var certificateArn string

	cmd := &cobra.Command{
		Use:   "delete-certificate",
		Short: "Delete an ACM certificate",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := acm.NewFromConfig(cfg, func(o *acm.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			_, err = client.DeleteCertificate(cmd.Context(), &acm.DeleteCertificateInput{
				CertificateArn: aws.String(certificateArn),
			})
			if err != nil {
				return fmt.Errorf("delete-certificate failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&certificateArn, "certificate-arn", "", "Certificate ARN")

	return cmd
}

func newACMGetCertificateCmd() *cobra.Command {
	var certificateArn string

	cmd := &cobra.Command{
		Use:   "get-certificate",
		Short: "Get an ACM certificate",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := acm.NewFromConfig(cfg, func(o *acm.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			out, err := client.GetCertificate(cmd.Context(), &acm.GetCertificateInput{
				CertificateArn: aws.String(certificateArn),
			})
			if err != nil {
				return fmt.Errorf("get-certificate failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&certificateArn, "certificate-arn", "", "Certificate ARN")

	return cmd
}

func newACMImportCertificateCmd() *cobra.Command {
	var certificate, privateKey, certificateChain, certificateArn string

	cmd := &cobra.Command{
		Use:   "import-certificate",
		Short: "Import an ACM certificate",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := newAWSConfig(cmd.Context())
			if err != nil {
				return err
			}

			client := acm.NewFromConfig(cfg, func(o *acm.Options) {
				o.BaseEndpoint = aws.String(endpointURL)
			})

			input := &acm.ImportCertificateInput{
				Certificate: []byte(certificate),
				PrivateKey:  []byte(privateKey),
			}

			if certificateChain != "" {
				input.CertificateChain = []byte(certificateChain)
			}

			if certificateArn != "" {
				input.CertificateArn = aws.String(certificateArn)
			}

			out, err := client.ImportCertificate(cmd.Context(), input)
			if err != nil {
				return fmt.Errorf("import-certificate failed: %w", err)
			}

			if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&certificate, "certificate", "", "Certificate body (PEM)")
	cmd.Flags().StringVar(&privateKey, "private-key", "", "Private key (PEM)")
	cmd.Flags().StringVar(&certificateChain, "certificate-chain", "", "Certificate chain (PEM)")
	cmd.Flags().StringVar(&certificateArn, "certificate-arn", "", "Certificate ARN (for reimport)")

	return cmd
}

package cloudfront

import (
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

const (
	xmlHeader       = `<?xml version="1.0" encoding="UTF-8"?>`
	cloudfrontXmlns = "http://cloudfront.amazonaws.com/doc/2020-05-31/"
)

// CreateDistribution handles the CreateDistribution operation.
func (s *Service) CreateDistribution(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeCloudFrontError(w, errMissingBody, "Request body is missing", http.StatusBadRequest)

		return
	}

	var req CreateDistributionRequest
	if err := xml.Unmarshal(body, &req); err != nil {
		writeCloudFrontError(w, errInvalidArgument, "Invalid request body", http.StatusBadRequest)

		return
	}

	dist, err := s.storage.CreateDistribution(r.Context(), &req)
	if err != nil {
		handleStorageError(w, err)

		return
	}

	resp := buildDistributionXML(dist)
	w.Header().Set("ETag", dist.ETag)
	w.Header().Set("Location", "/2020-05-31/distribution/"+dist.ID)
	writeXMLResponse(w, http.StatusCreated, resp)
}

// GetDistribution handles the GetDistribution operation.
func (s *Service) GetDistribution(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeCloudFrontError(w, errInvalidArgument, "Distribution ID is required", http.StatusBadRequest)

		return
	}

	dist, err := s.storage.GetDistribution(r.Context(), id)
	if err != nil {
		handleStorageError(w, err)

		return
	}

	resp := buildDistributionXML(dist)
	w.Header().Set("ETag", dist.ETag)
	writeXMLResponse(w, http.StatusOK, resp)
}

// ListDistributions handles the ListDistributions operation.
func (s *Service) ListDistributions(w http.ResponseWriter, r *http.Request) {
	marker := r.URL.Query().Get("Marker")
	maxItemsStr := r.URL.Query().Get("MaxItems")

	maxItems := 100
	if maxItemsStr != "" {
		if v, err := strconv.Atoi(maxItemsStr); err == nil && v > 0 {
			maxItems = v
		}
	}

	dists, nextMarker, err := s.storage.ListDistributions(r.Context(), marker, maxItems)
	if err != nil {
		handleStorageError(w, err)

		return
	}

	resp := buildDistributionListXML(dists, marker, maxItems, nextMarker)
	writeXMLResponse(w, http.StatusOK, resp)
}

// UpdateDistribution handles the UpdateDistribution operation.
func (s *Service) UpdateDistribution(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeCloudFrontError(w, errInvalidArgument, "Distribution ID is required", http.StatusBadRequest)

		return
	}

	etag := r.Header.Get("If-Match")
	if etag == "" {
		writeCloudFrontError(w, errPreconditionFailed, "The If-Match header is required", http.StatusPreconditionFailed)

		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeCloudFrontError(w, errMissingBody, "Request body is missing", http.StatusBadRequest)

		return
	}

	var req CreateDistributionRequest
	if err := xml.Unmarshal(body, &req); err != nil {
		writeCloudFrontError(w, errInvalidArgument, "Invalid request body", http.StatusBadRequest)

		return
	}

	dist, err := s.storage.UpdateDistribution(r.Context(), id, &req, etag)
	if err != nil {
		handleStorageError(w, err)

		return
	}

	resp := buildDistributionXML(dist)
	w.Header().Set("ETag", dist.ETag)
	writeXMLResponse(w, http.StatusOK, resp)
}

// DeleteDistribution handles the DeleteDistribution operation.
func (s *Service) DeleteDistribution(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeCloudFrontError(w, errInvalidArgument, "Distribution ID is required", http.StatusBadRequest)

		return
	}

	etag := r.Header.Get("If-Match")
	if etag == "" {
		writeCloudFrontError(w, errPreconditionFailed, "The If-Match header is required", http.StatusPreconditionFailed)

		return
	}

	if err := s.storage.DeleteDistribution(r.Context(), id, etag); err != nil {
		handleStorageError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateInvalidation handles the CreateInvalidation operation.
func (s *Service) CreateInvalidation(w http.ResponseWriter, r *http.Request) {
	distributionID := r.PathValue("id")
	if distributionID == "" {
		writeCloudFrontError(w, errInvalidArgument, "Distribution ID is required", http.StatusBadRequest)

		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeCloudFrontError(w, errMissingBody, "Request body is missing", http.StatusBadRequest)

		return
	}

	var req CreateInvalidationRequest
	if err := xml.Unmarshal(body, &req); err != nil {
		writeCloudFrontError(w, errInvalidArgument, "Invalid request body", http.StatusBadRequest)

		return
	}

	inv, err := s.storage.CreateInvalidation(r.Context(), distributionID, &req)
	if err != nil {
		handleStorageError(w, err)

		return
	}

	resp := buildInvalidationXML(inv)
	w.Header().Set("Location", "/2020-05-31/distribution/"+distributionID+"/invalidation/"+inv.ID)
	writeXMLResponse(w, http.StatusCreated, resp)
}

// GetInvalidation handles the GetInvalidation operation.
func (s *Service) GetInvalidation(w http.ResponseWriter, r *http.Request) {
	distributionID := r.PathValue("id")
	invalidationID := r.PathValue("invalidationId")

	if distributionID == "" || invalidationID == "" {
		writeCloudFrontError(w, errInvalidArgument, "Distribution ID and Invalidation ID are required", http.StatusBadRequest)

		return
	}

	inv, err := s.storage.GetInvalidation(r.Context(), distributionID, invalidationID)
	if err != nil {
		handleStorageError(w, err)

		return
	}

	resp := buildInvalidationXML(inv)
	writeXMLResponse(w, http.StatusOK, resp)
}

// GetDistributionConfig handles the GetDistributionConfig operation.
func (s *Service) GetDistributionConfig(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeCloudFrontError(w, errInvalidArgument, "Distribution ID is required", http.StatusBadRequest)

		return
	}

	dist, err := s.storage.GetDistribution(r.Context(), id)
	if err != nil {
		handleStorageError(w, err)

		return
	}

	resp := buildDistributionConfigXML(dist.DistributionConfig)
	w.Header().Set("ETag", dist.ETag)
	writeXMLResponse(w, http.StatusOK, resp)
}

// Helper functions.

func writeXMLResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("x-amzn-RequestId", uuid.New().String())
	w.WriteHeader(status)
	_, _ = io.WriteString(w, xmlHeader)
	_ = xml.NewEncoder(w).Encode(v)
}

func writeCloudFrontError(w http.ResponseWriter, code, message string, status int) {
	resp := ErrorResponse{
		Xmlns: cloudfrontXmlns,
		Error: ErrorDetail{
			Type:    "Sender",
			Code:    code,
			Message: message,
		},
		RequestID: uuid.New().String(),
	}
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("x-amzn-RequestId", resp.RequestID)
	w.WriteHeader(status)
	_, _ = io.WriteString(w, xmlHeader)
	_ = xml.NewEncoder(w).Encode(resp)
}

func handleStorageError(w http.ResponseWriter, err error) {
	var cfErr *Error
	if errors.As(err, &cfErr) {
		status := http.StatusBadRequest

		switch cfErr.Code {
		case errDistributionNotFound, errNoSuchInvalidation:
			status = http.StatusNotFound
		case errPreconditionFailed, errInvalidIfMatchVersion:
			status = http.StatusPreconditionFailed
		case errAccessDenied:
			status = http.StatusForbidden
		}

		writeCloudFrontError(w, cfErr.Code, cfErr.Message, status)

		return
	}

	writeCloudFrontError(w, "InternalError", "Internal server error", http.StatusInternalServerError)
}

func buildDistributionXML(dist *Distribution) *GetDistributionResult {
	return &GetDistributionResult{
		Xmlns:            cloudfrontXmlns,
		ID:               dist.ID,
		ARN:              dist.ARN,
		Status:           dist.Status,
		LastModifiedTime: dist.LastModifiedTime.Format(time.RFC3339),
		DomainName:       dist.DomainName,
		ActiveTrustedSigners: &ActiveTrustedSignersXML{
			Enabled:  dist.ActiveTrustedSigners.Enabled,
			Quantity: dist.ActiveTrustedSigners.Quantity,
		},
		ActiveTrustedKeyGroups: &ActiveTrustedKeyGroupsXML{
			Enabled:  dist.ActiveTrustedKeyGroups.Enabled,
			Quantity: dist.ActiveTrustedKeyGroups.Quantity,
		},
		DistributionConfig: buildDistributionConfigXML(dist.DistributionConfig),
	}
}

func buildDistributionConfigXML(config *DistributionConfig) *DistributionConfigXML {
	if config == nil {
		return nil
	}

	result := &DistributionConfigXML{
		CallerReference:   config.CallerReference,
		Comment:           config.Comment,
		Enabled:           config.Enabled,
		PriceClass:        config.PriceClass,
		DefaultRootObject: config.DefaultRootObject,
		HTTPVersion:       config.HTTPVersion,
		IsIPV6Enabled:     config.IsIPV6Enabled,
	}

	if config.Origins != nil {
		result.Origins = &OriginsXML{
			Quantity: config.Origins.Quantity,
		}
		if len(config.Origins.Items) > 0 {
			result.Origins.Items = &OriginList{}
			for _, o := range config.Origins.Items {
				origin := OriginXML{
					ID:                    o.ID,
					DomainName:            o.DomainName,
					OriginPath:            o.OriginPath,
					ConnectionAttempts:    o.ConnectionAttempts,
					ConnectionTimeout:     o.ConnectionTimeout,
					OriginAccessControlID: o.OriginAccessControlID,
				}
				if o.S3OriginConfig != nil {
					origin.S3OriginConfig = &S3OriginConfigXML{
						OriginAccessIdentity: o.S3OriginConfig.OriginAccessIdentity,
					}
				}
				if o.CustomOriginConfig != nil {
					origin.CustomOriginConfig = &CustomOriginConfigXML{
						HTTPPort:               o.CustomOriginConfig.HTTPPort,
						HTTPSPort:              o.CustomOriginConfig.HTTPSPort,
						OriginProtocolPolicy:   o.CustomOriginConfig.OriginProtocolPolicy,
						OriginReadTimeout:      o.CustomOriginConfig.OriginReadTimeout,
						OriginKeepaliveTimeout: o.CustomOriginConfig.OriginKeepaliveTimeout,
					}
					if o.CustomOriginConfig.OriginSSLProtocols != nil {
						origin.CustomOriginConfig.OriginSSLProtocols = &OriginSSLProtocolsXML{
							Quantity: o.CustomOriginConfig.OriginSSLProtocols.Quantity,
							Items:    o.CustomOriginConfig.OriginSSLProtocols.Items,
						}
					}
				}
				result.Origins.Items.Origin = append(result.Origins.Items.Origin, origin)
			}
		}
	}

	if config.DefaultCacheBehavior != nil {
		result.DefaultCacheBehavior = &DefaultCacheBehaviorXML{
			TargetOriginID:       config.DefaultCacheBehavior.TargetOriginID,
			ViewerProtocolPolicy: config.DefaultCacheBehavior.ViewerProtocolPolicy,
			MinTTL:               config.DefaultCacheBehavior.MinTTL,
			DefaultTTL:           config.DefaultCacheBehavior.DefaultTTL,
			MaxTTL:               config.DefaultCacheBehavior.MaxTTL,
			Compress:             config.DefaultCacheBehavior.Compress,
			CachePolicyID:        config.DefaultCacheBehavior.CachePolicyID,
		}

		if config.DefaultCacheBehavior.AllowedMethods != nil {
			result.DefaultCacheBehavior.AllowedMethods = &AllowedMethodsXML{
				Quantity: config.DefaultCacheBehavior.AllowedMethods.Quantity,
				Items:    config.DefaultCacheBehavior.AllowedMethods.Items,
			}
		}

		if config.DefaultCacheBehavior.ForwardedValues != nil {
			result.DefaultCacheBehavior.ForwardedValues = &ForwardedValuesXML{
				QueryString: config.DefaultCacheBehavior.ForwardedValues.QueryString,
			}
			if config.DefaultCacheBehavior.ForwardedValues.Cookies != nil {
				result.DefaultCacheBehavior.ForwardedValues.Cookies = &CookiesXML{
					Forward: config.DefaultCacheBehavior.ForwardedValues.Cookies.Forward,
				}
			}
			if config.DefaultCacheBehavior.ForwardedValues.Headers != nil {
				result.DefaultCacheBehavior.ForwardedValues.Headers = &HeadersXML{
					Quantity: config.DefaultCacheBehavior.ForwardedValues.Headers.Quantity,
					Items:    config.DefaultCacheBehavior.ForwardedValues.Headers.Items,
				}
			}
		}

		if config.DefaultCacheBehavior.TrustedSigners != nil {
			result.DefaultCacheBehavior.TrustedSigners = &TrustedSignersXML{
				Enabled:  config.DefaultCacheBehavior.TrustedSigners.Enabled,
				Quantity: config.DefaultCacheBehavior.TrustedSigners.Quantity,
				Items:    config.DefaultCacheBehavior.TrustedSigners.Items,
			}
		}

		if config.DefaultCacheBehavior.TrustedKeyGroups != nil {
			result.DefaultCacheBehavior.TrustedKeyGroups = &TrustedKeyGroupsXML{
				Enabled:  config.DefaultCacheBehavior.TrustedKeyGroups.Enabled,
				Quantity: config.DefaultCacheBehavior.TrustedKeyGroups.Quantity,
				Items:    config.DefaultCacheBehavior.TrustedKeyGroups.Items,
			}
		}
	}

	if config.Aliases != nil {
		result.Aliases = &AliasesXML{
			Quantity: config.Aliases.Quantity,
		}
		if len(config.Aliases.Items) > 0 {
			result.Aliases.Items = &ItemsXML{
				Items: config.Aliases.Items,
			}
		}
	}

	if config.ViewerCertificate != nil {
		result.ViewerCertificate = &ViewerCertificateXML{
			CloudFrontDefaultCertificate: config.ViewerCertificate.CloudFrontDefaultCertificate,
			IAMCertificateID:             config.ViewerCertificate.IAMCertificateID,
			ACMCertificateArn:            config.ViewerCertificate.ACMCertificateArn,
			SSLSupportMethod:             config.ViewerCertificate.SSLSupportMethod,
			MinimumProtocolVersion:       config.ViewerCertificate.MinimumProtocolVersion,
		}
	}

	return result
}

func buildDistributionListXML(dists []*Distribution, marker string, maxItems int, nextMarker string) *DistributionListXML {
	result := &DistributionListXML{
		Xmlns:       cloudfrontXmlns,
		Marker:      marker,
		MaxItems:    maxItems,
		IsTruncated: nextMarker != "",
		Quantity:    len(dists),
	}

	if len(dists) > 0 {
		result.Items = &DistributionSummaryList{}
		for _, d := range dists {
			summary := DistributionSummaryXML{
				ID:               d.ID,
				ARN:              d.ARN,
				Status:           d.Status,
				LastModifiedTime: d.LastModifiedTime.Format(time.RFC3339),
				DomainName:       d.DomainName,
				Enabled:          d.DistributionConfig.Enabled,
				Comment:          d.DistributionConfig.Comment,
				PriceClass:       d.DistributionConfig.PriceClass,
				HTTPVersion:      d.DistributionConfig.HTTPVersion,
				IsIPV6Enabled:    d.DistributionConfig.IsIPV6Enabled,
				CacheBehaviors:   &CacheBehaviorsXML{Quantity: 0},
			}

			if d.DistributionConfig.Aliases != nil {
				summary.Aliases = &AliasesXML{
					Quantity: d.DistributionConfig.Aliases.Quantity,
				}
			} else {
				summary.Aliases = &AliasesXML{Quantity: 0}
			}

			if d.DistributionConfig.Origins != nil {
				summary.Origins = &OriginsXML{
					Quantity: d.DistributionConfig.Origins.Quantity,
				}
				if len(d.DistributionConfig.Origins.Items) > 0 {
					summary.Origins.Items = &OriginList{}
					for _, o := range d.DistributionConfig.Origins.Items {
						origin := OriginXML{
							ID:         o.ID,
							DomainName: o.DomainName,
							OriginPath: o.OriginPath,
						}
						if o.S3OriginConfig != nil {
							origin.S3OriginConfig = &S3OriginConfigXML{
								OriginAccessIdentity: o.S3OriginConfig.OriginAccessIdentity,
							}
						}
						summary.Origins.Items.Origin = append(summary.Origins.Items.Origin, origin)
					}
				}
			}

			if d.DistributionConfig.DefaultCacheBehavior != nil {
				summary.DefaultCacheBehavior = &DefaultCacheBehaviorXML{
					TargetOriginID:       d.DistributionConfig.DefaultCacheBehavior.TargetOriginID,
					ViewerProtocolPolicy: d.DistributionConfig.DefaultCacheBehavior.ViewerProtocolPolicy,
				}
			}

			if d.DistributionConfig.ViewerCertificate != nil {
				summary.ViewerCertificate = &ViewerCertificateXML{
					CloudFrontDefaultCertificate: d.DistributionConfig.ViewerCertificate.CloudFrontDefaultCertificate,
					MinimumProtocolVersion:       d.DistributionConfig.ViewerCertificate.MinimumProtocolVersion,
				}
			}

			result.Items.DistributionSummary = append(result.Items.DistributionSummary, summary)
		}
	}

	if nextMarker != "" {
		result.NextMarker = nextMarker
	}

	return result
}

func buildInvalidationXML(inv *Invalidation) *InvalidationXML {
	result := &InvalidationXML{
		ID:         inv.ID,
		Status:     inv.Status,
		CreateTime: inv.CreateTime.Format(time.RFC3339),
	}

	if inv.InvalidationBatch != nil {
		result.InvalidationBatch = &InvalidationBatchXML{
			CallerReference: inv.InvalidationBatch.CallerReference,
		}
		if inv.InvalidationBatch.Paths != nil {
			result.InvalidationBatch.Paths = &PathsXML{
				Quantity: inv.InvalidationBatch.Paths.Quantity,
				Items:    inv.InvalidationBatch.Paths.Items,
			}
		}
	}

	return result
}

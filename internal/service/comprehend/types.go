// Package comprehend provides an in-memory implementation of AWS Comprehend.
package comprehend

// Error represents an error response.
type Error struct {
	Code    string `json:"__type"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// SentimentType represents the sentiment type.
type SentimentType string

// Sentiment types.
const (
	SentimentPositive SentimentType = "POSITIVE"
	SentimentNegative SentimentType = "NEGATIVE"
	SentimentNeutral  SentimentType = "NEUTRAL"
	SentimentMixed    SentimentType = "MIXED"
)

// EntityType represents the entity type.
type EntityType string

// Entity types.
const (
	EntityTypePerson       EntityType = "PERSON"
	EntityTypeLocation     EntityType = "LOCATION"
	EntityTypeOrganization EntityType = "ORGANIZATION"
	EntityTypeCommercial   EntityType = "COMMERCIAL_ITEM"
	EntityTypeEvent        EntityType = "EVENT"
	EntityTypeDate         EntityType = "DATE"
	EntityTypeQuantity     EntityType = "QUANTITY"
	EntityTypeTitle        EntityType = "TITLE"
	EntityTypeOther        EntityType = "OTHER"
)

// PiiEntityType represents the PII entity type.
type PiiEntityType string

// PII entity types.
const (
	PiiEntityTypeAddress      PiiEntityType = "ADDRESS"
	PiiEntityTypeAge          PiiEntityType = "AGE"
	PiiEntityTypeName         PiiEntityType = "NAME"
	PiiEntityTypeEmail        PiiEntityType = "EMAIL"
	PiiEntityTypePhone        PiiEntityType = "PHONE"
	PiiEntityTypeSSN          PiiEntityType = "SSN"
	PiiEntityTypeCreditCard   PiiEntityType = "CREDIT_DEBIT_NUMBER" //nolint:gosec // This is a PII type name, not a credential.
	PiiEntityTypeBankAccount  PiiEntityType = "BANK_ACCOUNT_NUMBER"
	PiiEntityTypePassport     PiiEntityType = "PASSPORT_NUMBER"
	PiiEntityTypeDriversID    PiiEntityType = "DRIVER_ID"
	PiiEntityTypeDateOfBirth  PiiEntityType = "DATE_TIME"
	PiiEntityTypeIPAddress    PiiEntityType = "IP_ADDRESS"
	PiiEntityTypeMACAddress   PiiEntityType = "MAC_ADDRESS"
	PiiEntityTypeURL          PiiEntityType = "URL"
	PiiEntityTypeUsername     PiiEntityType = "USERNAME"
	PiiEntityTypePassword     PiiEntityType = "PASSWORD"
	PiiEntityTypeAWSAccessKey PiiEntityType = "AWS_ACCESS_KEY"
	PiiEntityTypeAWSSecretKey PiiEntityType = "AWS_SECRET_KEY" //nolint:gosec // This is a PII type name, not a credential.
	PiiEntityTypeLicensePlate PiiEntityType = "LICENSE_PLATE"
	PiiEntityTypeVehicleID    PiiEntityType = "VEHICLE_IDENTIFICATION_NUMBER"
	PiiEntityTypeUKNIN        PiiEntityType = "UK_NATIONAL_INSURANCE_NUMBER"
	PiiEntityTypeCAHealthNum  PiiEntityType = "CA_HEALTH_NUMBER"
	PiiEntityTypeCASIN        PiiEntityType = "CA_SOCIAL_INSURANCE_NUMBER"
	PiiEntityTypeINAadhaar    PiiEntityType = "IN_AADHAAR"
	PiiEntityTypeINPAN        PiiEntityType = "IN_PERMANENT_ACCOUNT_NUMBER"
	PiiEntityTypeINVoterNum   PiiEntityType = "IN_VOTER_NUMBER"
)

// SyntaxTokenType represents the part of speech tag type.
type SyntaxTokenType string

// Syntax token types.
const (
	SyntaxTokenAdj   SyntaxTokenType = "ADJ"
	SyntaxTokenAdp   SyntaxTokenType = "ADP"
	SyntaxTokenAdv   SyntaxTokenType = "ADV"
	SyntaxTokenAux   SyntaxTokenType = "AUX"
	SyntaxTokenConj  SyntaxTokenType = "CONJ"
	SyntaxTokenCconj SyntaxTokenType = "CCONJ"
	SyntaxTokenDet   SyntaxTokenType = "DET"
	SyntaxTokenIntj  SyntaxTokenType = "INTJ"
	SyntaxTokenNoun  SyntaxTokenType = "NOUN"
	SyntaxTokenNum   SyntaxTokenType = "NUM"
	SyntaxTokenO     SyntaxTokenType = "O"
	SyntaxTokenPart  SyntaxTokenType = "PART"
	SyntaxTokenPron  SyntaxTokenType = "PRON"
	SyntaxTokenPropn SyntaxTokenType = "PROPN"
	SyntaxTokenPunct SyntaxTokenType = "PUNCT"
	SyntaxTokenSconj SyntaxTokenType = "SCONJ"
	SyntaxTokenSym   SyntaxTokenType = "SYM"
	SyntaxTokenVerb  SyntaxTokenType = "VERB"
)

// SentimentScore represents the confidence scores for each sentiment.
type SentimentScore struct {
	Positive float64 `json:"Positive"`
	Negative float64 `json:"Negative"`
	Neutral  float64 `json:"Neutral"`
	Mixed    float64 `json:"Mixed"`
}

// DominantLanguage represents a detected language.
type DominantLanguage struct {
	LanguageCode string  `json:"LanguageCode"`
	Score        float64 `json:"Score"`
}

// Entity represents a detected entity.
type Entity struct {
	BeginOffset int        `json:"BeginOffset"`
	EndOffset   int        `json:"EndOffset"`
	Score       float64    `json:"Score"`
	Text        string     `json:"Text"`
	Type        EntityType `json:"Type"`
}

// KeyPhrase represents a detected key phrase.
type KeyPhrase struct {
	BeginOffset int     `json:"BeginOffset"`
	EndOffset   int     `json:"EndOffset"`
	Score       float64 `json:"Score"`
	Text        string  `json:"Text"`
}

// PiiEntity represents a detected PII entity.
type PiiEntity struct {
	BeginOffset int           `json:"BeginOffset"`
	EndOffset   int           `json:"EndOffset"`
	Score       float64       `json:"Score"`
	Type        PiiEntityType `json:"Type"`
}

// SyntaxToken represents a syntax token.
type SyntaxToken struct {
	BeginOffset  int           `json:"BeginOffset"`
	EndOffset    int           `json:"EndOffset"`
	PartOfSpeech *PartOfSpeech `json:"PartOfSpeech,omitempty"`
	Text         string        `json:"Text"`
	TokenID      int           `json:"TokenId"`
}

// PartOfSpeech represents the part of speech.
type PartOfSpeech struct {
	Score float64         `json:"Score"`
	Tag   SyntaxTokenType `json:"Tag"`
}

// DetectSentimentRequest is the request for DetectSentiment.
type DetectSentimentRequest struct {
	LanguageCode string `json:"LanguageCode"`
	Text         string `json:"Text"`
}

// DetectSentimentResponse is the response for DetectSentiment.
type DetectSentimentResponse struct {
	Sentiment      SentimentType   `json:"Sentiment"`
	SentimentScore *SentimentScore `json:"SentimentScore"`
}

// DetectDominantLanguageRequest is the request for DetectDominantLanguage.
type DetectDominantLanguageRequest struct {
	Text string `json:"Text"`
}

// DetectDominantLanguageResponse is the response for DetectDominantLanguage.
type DetectDominantLanguageResponse struct {
	Languages []DominantLanguage `json:"Languages"`
}

// DetectEntitiesRequest is the request for DetectEntities.
type DetectEntitiesRequest struct {
	LanguageCode string `json:"LanguageCode"`
	Text         string `json:"Text"`
}

// DetectEntitiesResponse is the response for DetectEntities.
type DetectEntitiesResponse struct {
	Entities []Entity `json:"Entities"`
}

// DetectKeyPhrasesRequest is the request for DetectKeyPhrases.
type DetectKeyPhrasesRequest struct {
	LanguageCode string `json:"LanguageCode"`
	Text         string `json:"Text"`
}

// DetectKeyPhrasesResponse is the response for DetectKeyPhrases.
type DetectKeyPhrasesResponse struct {
	KeyPhrases []KeyPhrase `json:"KeyPhrases"`
}

// DetectPiiEntitiesRequest is the request for DetectPiiEntities.
type DetectPiiEntitiesRequest struct {
	LanguageCode string `json:"LanguageCode"`
	Text         string `json:"Text"`
}

// DetectPiiEntitiesResponse is the response for DetectPiiEntities.
type DetectPiiEntitiesResponse struct {
	Entities []PiiEntity `json:"Entities"`
}

// DetectSyntaxRequest is the request for DetectSyntax.
type DetectSyntaxRequest struct {
	LanguageCode string `json:"LanguageCode"`
	Text         string `json:"Text"`
}

// DetectSyntaxResponse is the response for DetectSyntax.
type DetectSyntaxResponse struct {
	SyntaxTokens []SyntaxToken `json:"SyntaxTokens"`
}

// BatchItemError represents an error for a batch item.
type BatchItemError struct {
	ErrorCode    string `json:"ErrorCode"`
	ErrorMessage string `json:"ErrorMessage"`
	Index        int    `json:"Index"`
}

// BatchDetectSentimentRequest is the request for BatchDetectSentiment.
type BatchDetectSentimentRequest struct {
	LanguageCode string   `json:"LanguageCode"`
	TextList     []string `json:"TextList"`
}

// BatchDetectSentimentItemResult represents a result item.
type BatchDetectSentimentItemResult struct {
	Index          int             `json:"Index"`
	Sentiment      SentimentType   `json:"Sentiment"`
	SentimentScore *SentimentScore `json:"SentimentScore"`
}

// BatchDetectSentimentResponse is the response for BatchDetectSentiment.
type BatchDetectSentimentResponse struct {
	ErrorList  []BatchItemError                 `json:"ErrorList"`
	ResultList []BatchDetectSentimentItemResult `json:"ResultList"`
}

// BatchDetectDominantLanguageRequest is the request for BatchDetectDominantLanguage.
type BatchDetectDominantLanguageRequest struct {
	TextList []string `json:"TextList"`
}

// BatchDetectDominantLanguageItemResult represents a result item.
type BatchDetectDominantLanguageItemResult struct {
	Index     int                `json:"Index"`
	Languages []DominantLanguage `json:"Languages"`
}

// BatchDetectDominantLanguageResponse is the response for BatchDetectDominantLanguage.
type BatchDetectDominantLanguageResponse struct {
	ErrorList  []BatchItemError                        `json:"ErrorList"`
	ResultList []BatchDetectDominantLanguageItemResult `json:"ResultList"`
}

// BatchDetectEntitiesRequest is the request for BatchDetectEntities.
type BatchDetectEntitiesRequest struct {
	LanguageCode string   `json:"LanguageCode"`
	TextList     []string `json:"TextList"`
}

// BatchDetectEntitiesItemResult represents a result item.
type BatchDetectEntitiesItemResult struct {
	Entities []Entity `json:"Entities"`
	Index    int      `json:"Index"`
}

// BatchDetectEntitiesResponse is the response for BatchDetectEntities.
type BatchDetectEntitiesResponse struct {
	ErrorList  []BatchItemError                `json:"ErrorList"`
	ResultList []BatchDetectEntitiesItemResult `json:"ResultList"`
}

// BatchDetectKeyPhrasesRequest is the request for BatchDetectKeyPhrases.
type BatchDetectKeyPhrasesRequest struct {
	LanguageCode string   `json:"LanguageCode"`
	TextList     []string `json:"TextList"`
}

// BatchDetectKeyPhrasesItemResult represents a result item.
type BatchDetectKeyPhrasesItemResult struct {
	Index      int         `json:"Index"`
	KeyPhrases []KeyPhrase `json:"KeyPhrases"`
}

// BatchDetectKeyPhrasesResponse is the response for BatchDetectKeyPhrases.
type BatchDetectKeyPhrasesResponse struct {
	ErrorList  []BatchItemError                  `json:"ErrorList"`
	ResultList []BatchDetectKeyPhrasesItemResult `json:"ResultList"`
}

// BatchDetectSyntaxRequest is the request for BatchDetectSyntax.
type BatchDetectSyntaxRequest struct {
	LanguageCode string   `json:"LanguageCode"`
	TextList     []string `json:"TextList"`
}

// BatchDetectSyntaxItemResult represents a result item.
type BatchDetectSyntaxItemResult struct {
	Index        int           `json:"Index"`
	SyntaxTokens []SyntaxToken `json:"SyntaxTokens"`
}

// BatchDetectSyntaxResponse is the response for BatchDetectSyntax.
type BatchDetectSyntaxResponse struct {
	ErrorList  []BatchItemError              `json:"ErrorList"`
	ResultList []BatchDetectSyntaxItemResult `json:"ResultList"`
}

// ContainsPiiEntitiesRequest is the request for ContainsPiiEntities.
type ContainsPiiEntitiesRequest struct {
	LanguageCode string `json:"LanguageCode"`
	Text         string `json:"Text"`
}

// EntityLabel represents an entity label.
type EntityLabel struct {
	Name  PiiEntityType `json:"Name"`
	Score float64       `json:"Score"`
}

// ContainsPiiEntitiesResponse is the response for ContainsPiiEntities.
type ContainsPiiEntitiesResponse struct {
	Labels []EntityLabel `json:"Labels"`
}

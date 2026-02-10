package detection

import (
	"context"
	"strings"

	"github.com/complyark/datalens/pkg/types"
)

// HeuristicStrategy detects PII by matching column names against a
// dictionary of 77+ known PII column name variations. This is the
// fastest strategy (no sample scanning) and provides useful signal
// even before any data is read.
//
// Base confidence: 0.70 (column name alone is suggestive, not definitive).
type HeuristicStrategy struct {
	// columnMap maps normalized column names to PII classification.
	columnMap map[string]heuristicMatch
}

type heuristicMatch struct {
	category    types.PIICategory
	piiType     types.PIIType
	sensitivity types.SensitivityLevel
}

// NewHeuristicStrategy creates a new column-name heuristic strategy.
func NewHeuristicStrategy() *HeuristicStrategy {
	return &HeuristicStrategy{
		columnMap: buildColumnMap(),
	}
}

func (s *HeuristicStrategy) Name() string                  { return "heuristic" }
func (s *HeuristicStrategy) Method() types.DetectionMethod { return types.DetectionMethodHeuristic }
func (s *HeuristicStrategy) Weight() float64               { return 0.70 }

// Detect checks whether the column name matches known PII patterns.
func (s *HeuristicStrategy) Detect(ctx context.Context, input Input) ([]Result, error) {
	normalized := normalizeColumnName(input.ColumnName)

	match, found := s.columnMap[normalized]
	if !found {
		return nil, nil
	}

	return []Result{
		{
			Category:    match.category,
			Type:        match.piiType,
			Sensitivity: match.sensitivity,
			Confidence:  0.70,
			Method:      types.DetectionMethodHeuristic,
			Reasoning:   "Column name '" + input.ColumnName + "' matches known PII pattern",
		},
	}, nil
}

// normalizeColumnName strips underscores, hyphens, and converts to lowercase.
func normalizeColumnName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "")
	name = strings.ReplaceAll(name, "-", "")
	name = strings.ReplaceAll(name, " ", "")
	return name
}

// buildColumnMap creates the comprehensive column name → PII mapping.
// Based on DataLens v1's 77+ column name heuristics.
func buildColumnMap() map[string]heuristicMatch {
	m := make(map[string]heuristicMatch)

	add := func(names []string, category types.PIICategory, piiType types.PIIType, sensitivity types.SensitivityLevel) {
		match := heuristicMatch{category: category, piiType: piiType, sensitivity: sensitivity}
		for _, n := range names {
			m[normalizeColumnName(n)] = match
		}
	}

	// Email patterns
	add([]string{
		"email", "e_mail", "emailaddress", "email_address", "mail",
		"emailid", "email_id", "user_email", "useremail",
	}, types.PIICategoryContact, types.PIITypeEmail, types.SensitivityMedium)

	// Phone patterns
	add([]string{
		"phone", "phonenumber", "phone_number", "mobile", "mobilenumber",
		"mobile_number", "cell", "cellphone", "telephone", "contact",
		"contactnumber", "contact_number", "tel", "fax",
	}, types.PIICategoryContact, types.PIITypePhone, types.SensitivityMedium)

	// Full name patterns
	add([]string{
		"name", "fullname", "full_name", "username", "user_name",
		"displayname", "display_name", "customername", "customer_name",
	}, types.PIICategoryIdentity, types.PIITypeName, types.SensitivityLow)

	// First name
	add([]string{
		"firstname", "first_name", "fname", "givenname", "given_name",
	}, types.PIICategoryIdentity, types.PIITypeName, types.SensitivityLow)

	// Last name
	add([]string{
		"lastname", "last_name", "lname", "surname", "familyname", "family_name",
	}, types.PIICategoryIdentity, types.PIITypeName, types.SensitivityLow)

	// Address patterns
	add([]string{
		"address", "street", "streetaddress", "street_address",
		"addr", "address1", "address2", "addressline1", "addressline2",
		"city", "state", "country",
	}, types.PIICategoryContact, types.PIITypeAddress, types.SensitivityMedium)

	// Postal/ZIP code
	add([]string{
		"postal", "postalcode", "postal_code", "zip", "zipcode", "zip_code",
		"pincode", "pin_code",
	}, types.PIICategoryContact, types.PIITypeAddress, types.SensitivityLow)

	// Aadhaar (India)
	add([]string{
		"aadhaar", "aadhar", "aadhaarnumber", "aadhaar_number",
		"aadhaarid", "aadhaar_id", "uid",
	}, types.PIICategoryGovernmentID, types.PIITypeAadhaar, types.SensitivityCritical)

	// PAN (India)
	add([]string{
		"pan", "pannumber", "pan_number", "pancard", "pan_card",
	}, types.PIICategoryGovernmentID, types.PIITypePAN, types.SensitivityHigh)

	// SSN (US)
	add([]string{
		"ssn", "socialsecurity", "social_security", "socialsecuritynumber",
		"social_security_number",
	}, types.PIICategoryGovernmentID, types.PIITypeSSN, types.SensitivityCritical)

	// Date of birth
	add([]string{
		"dob", "dateofbirth", "date_of_birth", "birthdate", "birth_date",
		"birthday",
	}, types.PIICategoryIdentity, types.PIITypeDOB, types.SensitivityMedium)

	// Credit card
	add([]string{
		"creditcard", "credit_card", "cardnumber", "card_number",
		"ccnumber", "cc_number", "ccn",
	}, types.PIICategoryFinancial, types.PIITypeCreditCard, types.SensitivityCritical)

	// Bank account
	add([]string{
		"bankaccount", "bank_account", "accountnumber", "account_number",
		"acctno", "acct_no", "iban", "ifsc",
	}, types.PIICategoryFinancial, types.PIITypeBankAccount, types.SensitivityCritical)

	// IP address
	add([]string{
		"ip", "ipaddress", "ip_address", "ipaddr", "ip_addr",
		"clientip", "client_ip", "remoteaddr", "remote_addr",
	}, types.PIICategoryBehavioral, types.PIITypeIPAddress, types.SensitivityLow)

	// Location
	add([]string{
		"location", "latitude", "longitude", "lat", "lng", "lon",
		"geolocation", "geo_location", "coordinates",
	}, types.PIICategoryLocation, types.PIITypeAddress, types.SensitivityMedium)

	// Gender
	add([]string{
		"gender", "sex",
	}, types.PIICategoryIdentity, types.PIITypeGender, types.SensitivityLow)

	// Passport
	add([]string{
		"passport", "passportnumber", "passport_number", "passportno", "passport_no",
	}, types.PIICategoryGovernmentID, types.PIITypePassport, types.SensitivityHigh)

	// National ID (generic)
	add([]string{
		"nationalid", "national_id", "idnumber", "id_number",
		"governmentid", "government_id",
	}, types.PIICategoryGovernmentID, types.PIITypeNationalID, types.SensitivityHigh)

	return m
}

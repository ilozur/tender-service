package models

func ValidateTenderServiceType(t TenderServiceType) bool {
	values := []TenderServiceType{Construction, Delivery, Manufacture}

	for _, el := range values {
		if el == t {
			return true
		}
	}
	return false
}

func ValidateTenderStatus(t TenderStatus) bool {
	values := []TenderStatus{TenderCreated, TenderPublished, TenderClosed}

	for _, el := range values {
		if el == t {
			return true
		}
	}
	return false
}

func ValidateOrganizationType(t OrganizationType) bool {
	values := []OrganizationType{IEOrganization, LLCOrganization, JSCOrganization}

	for _, el := range values {
		if el == t {
			return true
		}
	}
	return false
}

func ValidateBidStatus(t BidStatus) bool {
	values := []BidStatus{BidCreated, BidPublished, BidCanceled}

	for _, el := range values {
		if el == t {
			return true
		}
	}
	return false
}

func ValidateBidDecision(t BidStatus) bool {
	values := []BidStatus{BidRejected, BidApproved}

	for _, el := range values {
		if el == t {
			return true
		}
	}
	return false
}

func ValidateBidAuthorType(t BidAuthorType) bool {
	values := []BidAuthorType{BidAuthorOrganization, BidAuthorUser}

	for _, el := range values {
		if el == t {
			return true
		}
	}
	return false
}

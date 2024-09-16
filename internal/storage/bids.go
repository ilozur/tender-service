package storage

import (
	"errors"
	"tender_service/internal/handlers/bids/bid_feedback"
	"tender_service/internal/handlers/bids/bid_submit_decision"
	"tender_service/internal/handlers/bids/bids_rollback"
	"tender_service/internal/handlers/bids/get_bid_status"
	"tender_service/internal/handlers/bids/get_bids"
	"tender_service/internal/handlers/bids/get_my_bids"
	"tender_service/internal/handlers/bids/get_reviews"
	"tender_service/internal/handlers/bids/new"
	"tender_service/internal/handlers/bids/patch_bid"
	"tender_service/internal/handlers/bids/put_bid_status"
	"tender_service/internal/lib/response"
	"tender_service/internal/storage/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"tender_service/internal/lib/time_converter"
)

func (s *Storage) SaveBid(req newbid.Request) (newbid.Response, error) {
	user, err := s.GetUserById(req.AuthorID)
	if err != nil {
		return newbid.Response{}, err
	}

	orgID, err := s.GetOrganization(user.ID)
	if err != nil {
		return newbid.Response{}, err
	}

	tender, err := s.GetTender(req.TenderId)

	if err != nil {
		return newbid.Response{}, err
	}

	if tender.Status != models.TenderPublished {
		return newbid.Response{}, response.ErrNoRights
	}

	newBid := models.Bid{
		Name:             req.Name,
		Description:      req.Description,
		AuthorType:       models.BidAuthorUser,
		Status:           models.BidCreated,
		TenderID:         req.TenderId,
		EmployeeUsername: user.Username,
		OrganizationID:   orgID,
	}

	result := s.db.Create(&newBid)

	if result.Error != nil {
		return newbid.Response{}, response.ErrInternalError
	}

	newBidVersion := models.BidVersion{
		Name:             req.Name,
		Description:      req.Description,
		BidID:            newBid.ID,
		AuthorType:       models.BidAuthorUser,
		Status:           models.BidCreated,
		TenderID:         req.TenderId,
		EmployeeUsername: user.Username,
		OrganizationID:   orgID,
	}

	result = s.db.Create(&newBidVersion)
	if result.Error != nil {
		return newbid.Response{}, response.ErrInternalError
	}

	return newbid.Response{
		ID:          newBid.ID,
		Version:     uint(newBid.Version),
		CreatedAt:   time_converter.Time(newBid.CreatedAt),
		Name:        newBid.Name,
		Description: newBid.Description,
		AuthorType:  string(newBid.AuthorType),
		Status:      string(newBid.Status),
		AuthorID:    user.ID,
	}, nil
}

func (s *Storage) BidStatus(req getbidstatus.Request) (getbidstatus.Response, error) {

	_, err := s.GetUser(req.UserName)
	if err != nil {
		return getbidstatus.Response{}, err
	}

	var bid models.Bid
	query := s.db.Model(&models.Bid{})

	result := query.Where("id = ?", req.BidID).First(&bid, req.BidID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return getbidstatus.Response{}, response.ErrBidNotExists
		}
		return getbidstatus.Response{}, response.ErrInternalError
	}

	if bid.EmployeeUsername != req.UserName {
		return getbidstatus.Response{}, response.ErrNoRights
	}

	return getbidstatus.Response{
		Status: string(bid.Status),
	}, nil
}

func (s *Storage) BidSubmitDecision(req bidsubmitdecision.Request) (bidsubmitdecision.Response, error) {
	user, err := s.GetUser(req.UserName)
	if err != nil {
		return bidsubmitdecision.Response{}, err
	}

	var bid models.Bid
	query := s.db.Model(&models.Bid{})

	result := query.Where("status = ?", models.BidPublished).First(&bid, req.BidID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return bidsubmitdecision.Response{}, response.ErrBidNotExists
		}
		return bidsubmitdecision.Response{}, response.ErrInternalError
	}

	tender, err := s.GetTender(bid.TenderID)

	if err != nil {
		return bidsubmitdecision.Response{}, err
	}

	orgID, err := s.GetOrganization(user.ID)

	if err != nil {
		if errors.Is(err, response.ErrUserNotExists) {
			return bidsubmitdecision.Response{}, response.ErrNoRights
		}
		return bidsubmitdecision.Response{}, response.ErrInternalError
	}

	if tender.OrganizationID != orgID {
		return bidsubmitdecision.Response{}, response.ErrNoRights
	}

	bid.Status = models.BidStatus(req.Decision)

	err = s.UpdateBid(&bid)

	if err != nil {
		return bidsubmitdecision.Response{}, err
	}

	tender.Status = models.TenderClosed

	err = s.UpdateTender(tender)

	if err != nil {
		return bidsubmitdecision.Response{}, err
	}

	return bidsubmitdecision.Response{
		ID:          bid.ID,
		Version:     uint(bid.Version),
		CreatedAt:   time_converter.Time(bid.CreatedAt),
		Name:        bid.Name,
		Description: bid.Description,
		AuthorType:  string(bid.AuthorType),
		Status:      string(bid.Status),
		AuthorID:    user.ID,
	}, nil
}

func (s *Storage) BidFeedback(req bidfeedback.Request) (bidfeedback.Response, error) {
	_, err := s.GetUser(req.UserName)
	if err != nil {
		return bidfeedback.Response{}, err
	}

	var bid models.Bid
	query := s.db.Model(&models.Bid{})

	result := query.Where("status = ?", models.BidPublished).First(&bid, req.BidID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return bidfeedback.Response{}, response.ErrBidNotExists
		}
		return bidfeedback.Response{}, response.ErrInternalError
	}

	tender, err := s.GetTender(bid.TenderID)

	if err != nil {
		return bidfeedback.Response{}, err
	}

	user, err := s.GetUser(tender.EmployeeUsername)
	if err != nil {
		return bidfeedback.Response{}, err
	}

	orgID, err := s.GetOrganization(user.ID)

	if err != nil {
		if errors.Is(err, response.ErrUserNotExists) {
			return bidfeedback.Response{}, response.ErrNoRights
		}
		return bidfeedback.Response{}, response.ErrInternalError
	}

	if tender.OrganizationID != orgID {
		return bidfeedback.Response{}, response.ErrNoRights
	}

	err = s.CreateBidFeedback(req.BidID, req.BidFeedback, req.UserName, orgID)

	if err != nil {
		return bidfeedback.Response{}, err
	}

	return bidfeedback.Response{
		ID:          bid.ID,
		Version:     uint(bid.Version),
		CreatedAt:   time_converter.Time(bid.CreatedAt),
		Name:        bid.Name,
		Description: bid.Description,
		AuthorType:  string(bid.AuthorType),
		Status:      string(bid.Status),
		AuthorID:    user.ID,
	}, nil
}

func (s *Storage) CreateBidFeedback(bidID uuid.UUID, feedback string, username string, orgID uuid.UUID) error {
	res := s.db.Create(&models.BidFeedback{
		Feedback:         feedback,
		BidID:            bidID,
		EmployeeUsername: username,
		OrganizationID:   orgID,
	})

	if res.Error != nil {
		return response.ErrInternalError
	}
	return nil
}

func (s *Storage) BidStatusPutter(req putbidstatus.Request) (putbidstatus.Response, error) {
	user, err := s.GetUser(req.UserName)
	if err != nil {
		return putbidstatus.Response{}, err
	}

	orgID, err := s.GetOrganization(user.ID)

	if err != nil {
		return putbidstatus.Response{}, err
	}

	var bid models.Bid
	query := s.db.Model(&models.Bid{})

	result := query.First(&bid, req.BidID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return putbidstatus.Response{}, response.ErrBidNotExists
		}
		return putbidstatus.Response{}, response.ErrInternalError
	}

	if orgID != (bid.OrganizationID) {
		return putbidstatus.Response{}, response.ErrNoRights
	}

	bid.Status = models.BidStatus(req.Status)

	err = s.UpdateBid(&bid)

	if err != nil {
		return putbidstatus.Response{}, response.ErrInternalError
	}

	return putbidstatus.Response{
		ID:          bid.ID,
		Version:     uint(bid.Version),
		CreatedAt:   time_converter.Time(bid.CreatedAt),
		Name:        bid.Name,
		Description: bid.Description,
		AuthorType:  string(bid.AuthorType),
		Status:      string(bid.Status),
		AuthorID:    user.ID,
	}, nil
}

func (s *Storage) PatchBid(req patchbid.Request) (patchbid.Response, error) {
	user, err := s.GetUser(req.UserName)
	if err != nil {
		return patchbid.Response{}, err
	}

	orgID, err := s.GetOrganization(user.ID)

	if err != nil {
		return patchbid.Response{}, err
	}

	var bid models.Bid
	query := s.db.Model(&models.Bid{})
	result := query.First(&bid, req.BidID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return patchbid.Response{}, response.ErrBidNotExists
		}
		return patchbid.Response{}, response.ErrInternalError
	}

	if req.TenderID != uuid.Nil {
		_, err = s.GetTender(req.TenderID)

		if err != nil {
			return patchbid.Response{}, err
		}
	}

	if orgID != (bid.OrganizationID) {
		return patchbid.Response{}, response.ErrNoRights
	}

	PatchBid(&bid, req)

	err = s.UpdateBid(&bid)

	if err != nil {
		return patchbid.Response{}, response.ErrInternalError
	}

	return patchbid.Response{
		ID:          bid.ID,
		Version:     uint(bid.Version),
		CreatedAt:   time_converter.Time(bid.CreatedAt),
		Name:        bid.Name,
		Description: bid.Description,
		AuthorType:  string(bid.AuthorType),
		Status:      string(bid.Status),
		AuthorID:    user.ID,
	}, nil
}

func (s *Storage) BidRollback(req bidsrollback.Request) (bidsrollback.Response, error) {
	user, err := s.GetUser(req.UserName)
	if err != nil {
		return bidsrollback.Response{}, err
	}

	orgID, err := s.GetOrganization(user.ID)

	if err != nil {
		return bidsrollback.Response{}, err
	}

	var bidVersion models.BidVersion

	query := s.db.Model(&models.BidVersion{})
	query = query.Where("version = ? AND bid_id = ? ", req.Version, req.BidID)
	result := query.First(&bidVersion)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return bidsrollback.Response{}, response.ErrBidNotExists
		}
		return bidsrollback.Response{}, response.ErrInternalError
	}

	if bidVersion.OrganizationID != orgID {
		return bidsrollback.Response{}, response.ErrNoRights
	}

	var bid models.Bid
	query = s.db.Model(&models.Bid{})
	query = query.Where("id = ? ", bidVersion.BidID)
	result = query.First(&bid)

	if result.Error != nil {
		return bidsrollback.Response{}, response.ErrInternalError
	}

	s.UpdateBidByVersion(&bid, &bidVersion)

	s.UpdateBid(&bid)

	return bidsrollback.Response{
		ID:          bid.ID,
		Version:     uint(bid.Version),
		CreatedAt:   time_converter.Time(bid.CreatedAt),
		Name:        bid.Name,
		Description: bid.Description,
		AuthorType:  string(bid.AuthorType),
		Status:      string(bid.Status),
		AuthorID:    user.ID,
	}, nil
}

func (s *Storage) UpdateBidByVersion(bid *models.Bid, newBid *models.BidVersion) {
	bid.Name = newBid.Name
	bid.Description = newBid.Description
	bid.TenderID = newBid.TenderID
	bid.Status = newBid.Status
	bid.AuthorType = newBid.AuthorType
}

func (s *Storage) GetBids(req getbids.Request) (getbids.ResponseList, error) {
	usr, err := s.GetUser(req.Username)
	if err != nil {
		return getbids.ResponseList{}, err
	}

	orgID, err := s.GetOrganization(usr.ID)

	if err != nil {
		return getbids.ResponseList{}, err
	}

	tender, err := s.GetTender(req.TenderID)

	if err != nil {
		return getbids.ResponseList{}, err
	}

	if tender.OrganizationID != orgID {
		return getbids.ResponseList{}, response.ErrNoRights
	}

	var bids []models.Bid
	query := s.db.Model(&models.Bid{})
	query = query.Where("tender_id = ? AND status = ?", req.TenderID, string(models.BidPublished)).Limit(int(req.Limit)).Offset(int(req.OffSet))

	result := query.Find(&bids)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return getbids.ResponseList{}, response.ErrInternalError
	}

	var responses []getbids.Response

	for _, el := range bids {
		res := getbids.Response{
			ID:          el.ID,
			Version:     uint(el.Version),
			CreatedAt:   time_converter.Time(el.CreatedAt),
			Name:        el.Name,
			AuthorType:  string(el.AuthorType),
			AuthorID:    usr.ID,
			Description: el.Description,
			Status:      string(el.Status),
		}
		responses = append(responses, res)
	}

	return getbids.ResponseList{
		Response: responses,
	}, nil
}

func (s *Storage) GetReviews(req getreviews.Request) (getreviews.ResponseList, error) {
	usrRequester, err := s.GetUser(req.RequesterUsername)
	if err != nil {
		return getreviews.ResponseList{}, err
	}

	_, err = s.GetUser(req.AuthorUsername)
	if err != nil {
		return getreviews.ResponseList{}, err
	}

	tender, err := s.GetTender(req.TenderID)

	if err != nil {
		return getreviews.ResponseList{}, err
	}

	orgID, err := s.GetOrganization(usrRequester.ID)

	if err != nil {
		return getreviews.ResponseList{}, err
	}

	if tender.OrganizationID != orgID {
		return getreviews.ResponseList{}, response.ErrNoRights
	}

	var bids []models.Bid
	query := s.db.Model(&models.Bid{})
	query = query.Where("tender_id = ? AND employee_username = ?", req.TenderID, req.AuthorUsername).Limit(int(req.Limit)).Offset(int(req.OffSet))

	result := query.Find(&bids)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return getreviews.ResponseList{}, response.ErrInternalError
	}

	var bidsID []uuid.UUID

	for _, el := range bids {
		bidsID = append(bidsID, el.ID)
	}
	var bidsFeedback []models.BidFeedback

	query = s.db.Model(&models.BidFeedback{})
	query = query.Where("bid_id IN ?", bidsID)
	result = query.Find(&bidsFeedback)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return getreviews.ResponseList{}, result.Error
	}

	var responses []getreviews.Response

	for _, el := range bidsFeedback {
		res := getreviews.Response{
			ID:          el.ID,
			CreatedAt:   time_converter.Time(el.CreatedAt),
			Description: el.Feedback,
		}
		responses = append(responses, res)
	}

	return getreviews.ResponseList{
		Response: responses,
	}, nil
}

func (s *Storage) GetMyBids(req getmybids.Request) (getmybids.ResponseList, error) {
	usr, err := s.GetUser(req.UserName)
	if err != nil {
		return getmybids.ResponseList{}, err
	}

	var bids []models.Bid
	query := s.db.Model(&models.Bid{})
	query = query.Where("employee_username = ?", req.UserName).Limit(int(req.Limit)).Offset(int(req.OffSet))

	result := query.Find(&bids)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return getmybids.ResponseList{}, response.ErrInternalError
	}

	var responses []getmybids.Response

	for _, el := range bids {
		res := getmybids.Response{
			ID:          el.ID,
			Version:     uint(el.Version),
			CreatedAt:   time_converter.Time(el.CreatedAt),
			Name:        el.Name,
			AuthorType:  string(el.AuthorType),
			AuthorID:    usr.ID,
			Description: el.Description,
			Status:      string(el.Status),
		}
		responses = append(responses, res)
	}

	return getmybids.ResponseList{
		Response: responses,
	}, nil
}

func PatchBid(bid *models.Bid, values patchbid.Request) {
	if values.Description != "" {
		bid.Description = values.Description
	}

	if values.Name != "" {
		bid.Name = values.Name
	}

	if values.Status != "" {
		bid.Status = models.BidStatus(values.Status)
	}

	if values.Status != "" {
		bid.Status = models.BidStatus(values.Status)
	}

	if values.TenderID != uuid.Nil {
		bid.TenderID = values.TenderID
	}
}

func (s *Storage) UpdateBid(bid *models.Bid) error {
	bid.Version++

	if err := s.db.Create(&models.BidVersion{
		Name:             bid.Name,
		Description:      bid.Description,
		BidID:            bid.ID,
		AuthorType:       models.BidAuthorUser,
		Status:           models.BidStatus(bid.Status),
		TenderID:         bid.TenderID,
		EmployeeUsername: bid.EmployeeUsername,
		OrganizationID:   bid.OrganizationID,
		Version:          bid.Version,
	}).Error; err != nil {
		return response.ErrInternalError
	}

	result := s.db.Save(&bid)
	if result.Error != nil {
		return response.ErrInternalError
	}

	return nil
}

func (s *Storage) GetTender(tenderID uuid.UUID) (*models.Tender, error) {
	if tenderID == uuid.Nil {
		return &models.Tender{}, response.ErrTenderNotExists
	}
	var tender models.Tender
	query := s.db.Model(&models.Tender{})
	result := query.Where("id = ?", tenderID).First(&tender)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return &models.Tender{}, response.ErrTenderNotExists
		}
		return &models.Tender{}, response.ErrInternalError
	}
	return &tender, nil
}

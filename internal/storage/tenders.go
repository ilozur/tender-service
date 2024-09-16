package storage

import (
	"log/slog"
	getmytenders "tender_service/internal/handlers/tenders/get_my_tenders"
	gettenderstatus "tender_service/internal/handlers/tenders/get_tender_status"
	gettenders "tender_service/internal/handlers/tenders/get_tenders"
	newtender "tender_service/internal/handlers/tenders/new_tender"
	patchtenderstatus "tender_service/internal/handlers/tenders/patch_tender_status"
	puttenderstatus "tender_service/internal/handlers/tenders/put_tender_status"
	tendersrollback "tender_service/internal/handlers/tenders/tenders_rollback"
	"tender_service/internal/lib/response"
	"tender_service/internal/storage/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"tender_service/internal/lib/time_converter"
)

func (s *Storage) GetUser(userName string) (*models.Employee, error) {
	if userName == "" {
		return &models.Employee{}, response.ErrUserNotExists
	}
	var user models.Employee
	query := s.db.Model(&models.Employee{})
	result := query.Where("username = ?", userName).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return &models.Employee{}, response.ErrUserNotExists
		}
		return &models.Employee{}, response.ErrInternalError
	}
	return &user, nil
}

func (s *Storage) GetUserById(userID uuid.UUID) (*models.Employee, error) {
	if userID == uuid.Nil {
		return &models.Employee{}, response.ErrUserNotExists
	}
	var user models.Employee
	query := s.db.Model(&models.Employee{})
	result := query.Where("id = ?", userID).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return &models.Employee{}, response.ErrUserNotExists
		}
		return &models.Employee{}, response.ErrInternalError
	}
	return &user, nil
}

func (s *Storage) GetOrganization(userID uuid.UUID) (uuid.UUID, error) {
	if userID == uuid.Nil {
		return uuid.Nil, response.ErrUserNotExists
	}
	var organization models.OrganizationResponsible
	query := s.db.Model(&models.OrganizationResponsible{})
	result := query.Where("user_id = ?", userID).First(&organization)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return uuid.Nil, response.ErrUserNotExists
		}
		return uuid.Nil, response.ErrInternalError
	}
	return organization.OrganizationID, nil
}

func (s *Storage) SaveTender(req newtender.Request) (newtender.Response, error) {
	user, err := s.GetUser(req.CreatorUsername)
	if err != nil {
		return newtender.Response{}, err
	}

	var orgUser models.OrganizationResponsible
	query := s.db.Model(&models.OrganizationResponsible{})

	result := query.Where("organization_id = ? AND user_id = ?", req.OrganizationId, user.ID).First(&orgUser)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return newtender.Response{}, response.ErrNoRights
		}
		return newtender.Response{}, response.ErrInternalError
	}

	newTender := models.Tender{Name: req.Name, Description: req.Description, ServiceType: models.TenderServiceType(req.ServiceType), Status: models.TenderCreated, EmployeeUsername: user.Username, OrganizationID: req.OrganizationId}

	result = s.db.Create(&newTender)

	if result.Error != nil {
		return newtender.Response{}, response.ErrInternalError
	}

	newTenderVersion := models.TenderVersion{TenderID: newTender.ID, Name: newTender.Name, Description: newTender.Description, ServiceType: newTender.ServiceType, Status: newTender.Status, EmployeeUsername: newTender.EmployeeUsername, OrganizationID: newTender.OrganizationID}

	slog.Info("start")

	result = s.db.Create(&newTenderVersion)
	if result.Error != nil {
		return newtender.Response{}, response.ErrInternalError
	}
	slog.Info("end")

	return newtender.Response{
		ID:          newTender.ID,
		Version:     newTender.Version,
		CreatedAt:   time_converter.Time(newTender.CreatedAt),
		Name:        newTender.Name,
		Description: newTender.Description,
		ServiceType: string(newTender.ServiceType),
		Status:      string(newTender.Status),
	}, nil
}

func (s *Storage) Status(req gettenderstatus.Request) (gettenderstatus.Response, error) {

	_, err := s.GetUser(req.UserName)
	if err != nil {
		return gettenderstatus.Response{}, err
	}
	var tender models.Tender
	query := s.db.Model(&models.Tender{})

	result := query.Where("id = ?", req.TenderID).First(&tender, req.TenderID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return gettenderstatus.Response{}, response.ErrTenderNotExists
		}
		return gettenderstatus.Response{}, response.ErrInternalError
	}

	if tender.EmployeeUsername != req.UserName {
		return gettenderstatus.Response{}, response.ErrNoRights
	}

	return gettenderstatus.Response{
		Status: string(tender.Status),
	}, nil
}

func (s *Storage) StatusPut(req puttenderstatus.Request) (puttenderstatus.Response, error) {
	user, err := s.GetUser(req.UserName)
	if err != nil {
		return puttenderstatus.Response{}, err
	}

	orgID, err := s.GetOrganization(user.ID)

	if err != nil {
		return puttenderstatus.Response{}, err
	}

	var tender models.Tender
	query := s.db.Model(&models.Tender{})

	result := query.First(&tender, req.TenderID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return puttenderstatus.Response{}, response.ErrTenderNotExists
		}
		return puttenderstatus.Response{}, response.ErrInternalError
	}

	if orgID != (tender.OrganizationID) {
		return puttenderstatus.Response{}, response.ErrNoRights
	}

	tender.Status = models.TenderStatus(req.Status)

	err = s.UpdateTender(&tender)

	if err != nil {
		return puttenderstatus.Response{}, response.ErrInternalError
	}

	return puttenderstatus.Response{
		ID:          tender.ID,
		Version:     tender.Version,
		CreatedAt:   time_converter.Time(tender.CreatedAt),
		Name:        tender.Name,
		Description: tender.Description,
		ServiceType: string(tender.ServiceType),
		Status:      string(tender.Status),
	}, nil
}

func (s *Storage) PatchTender(req patchtenderstatus.Request) (patchtenderstatus.Response, error) {
	user, err := s.GetUser(req.UserName)
	if err != nil {
		return patchtenderstatus.Response{}, err
	}

	orgID, err := s.GetOrganization(user.ID)

	if err != nil {
		return patchtenderstatus.Response{}, err
	}

	var tender models.Tender
	query := s.db.Model(&models.Tender{})
	result := query.First(&tender, req.TenderID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return patchtenderstatus.Response{}, response.ErrTenderNotExists
		}
		return patchtenderstatus.Response{}, response.ErrInternalError
	}

	if orgID != (tender.OrganizationID) {
		return patchtenderstatus.Response{}, response.ErrNoRights
	}

	PatchTender(&tender, req)

	err = s.UpdateTender(&tender)

	if err != nil {
		return patchtenderstatus.Response{}, response.ErrInternalError
	}

	return patchtenderstatus.Response{
		ID:          tender.ID,
		Version:     tender.Version,
		CreatedAt:   time_converter.Time(tender.CreatedAt),
		Name:        tender.Name,
		Description: tender.Description,
		ServiceType: string(tender.ServiceType),
		Status:      string(tender.Status),
	}, nil
}

func (s *Storage) GetTenders(req gettenders.Request) (gettenders.ResponseList, error) {
	var tenders []models.Tender
	var result *gorm.DB
	query := s.db.Model(&models.Tender{})
	if len(req.SeviceType) == 0 {
		query = query.Where("status = ?", string(models.TenderPublished)).Limit(int(req.Limit)).Offset(int(req.OffSet))
	} else {
		query = query.Where("status = ? AND service_type IN ?", string(models.TenderPublished), req.SeviceType).Limit(int(req.Limit)).Offset(int(req.OffSet))
	}

	result = query.Find(&tenders)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return gettenders.ResponseList{}, response.ErrInternalError
	}

	var responses []gettenders.Response

	for _, el := range tenders {
		res := gettenders.Response{
			ID:          el.ID,
			Version:     el.Version,
			CreatedAt:   time_converter.Time(el.CreatedAt),
			Name:        el.Name,
			Description: el.Description,
			ServiceType: string(el.ServiceType),
			Status:      string(el.Status),
		}
		responses = append(responses, res)
	}

	return gettenders.ResponseList{
		Response: responses,
	}, nil
}

func (s *Storage) GetMyTenders(req getmytenders.Request) (getmytenders.ResponseList, error) {
	_, err := s.GetUser(req.UserName)
	if err != nil {
		return getmytenders.ResponseList{}, err
	}

	var tenders []models.Tender
	query := s.db.Model(&models.Tender{})
	query = query.Where("employee_username = ?", req.UserName).Limit(int(req.Limit)).Offset(int(req.OffSet))

	result := query.Find(&tenders)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return getmytenders.ResponseList{}, response.ErrInternalError
	}

	var responses []getmytenders.Response

	for _, el := range tenders {
		res := getmytenders.Response{
			ID:          el.ID,
			Version:     el.Version,
			CreatedAt:   time_converter.Time(el.CreatedAt),
			Name:        el.Name,
			Description: el.Description,
			ServiceType: string(el.ServiceType),
			Status:      string(el.Status),
		}
		responses = append(responses, res)
	}

	return getmytenders.ResponseList{
		Response: responses,
	}, nil
}

func (s *Storage) TenderRollback(req tendersrollback.Request) (tendersrollback.Response, error) {
	user, err := s.GetUser(req.UserName)
	if err != nil {
		return tendersrollback.Response{}, err
	}

	orgID, err := s.GetOrganization(user.ID)

	if err != nil {
		return tendersrollback.Response{}, err
	}

	var tenderVersion models.TenderVersion

	query := s.db.Model(&models.TenderVersion{})
	query = query.Where("version = ? AND tender_id = ? ", req.Version, req.TenderID)
	result := query.First(&tenderVersion)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return tendersrollback.Response{}, response.ErrTenderNotExists
		}
		return tendersrollback.Response{}, response.ErrInternalError
	}

	if tenderVersion.OrganizationID != orgID {
		return tendersrollback.Response{}, response.ErrNoRights
	}

	var tender models.Tender
	query = s.db.Model(&models.Tender{})
	query = query.Where("id = ? ", tenderVersion.TenderID)
	result = query.First(&tender)

	if result.Error != nil {
		return tendersrollback.Response{}, response.ErrInternalError
	}

	s.UpdateTenderByVersion(&tender, &tenderVersion)

	s.UpdateTender(&tender)

	return tendersrollback.Response{
		ID:          tender.ID,
		Version:     tender.Version,
		CreatedAt:   time_converter.Time(tender.CreatedAt),
		Name:        tender.Name,
		Description: tender.Description,
		ServiceType: string(tender.ServiceType),
		Status:      string(tender.Status),
	}, nil
}

func PatchTender(tender *models.Tender, values patchtenderstatus.Request) {
	if values.Description != "" {
		tender.Description = values.Description
	}

	if values.Name != "" {
		tender.Name = values.Name
	}

	if values.ServiceType != "" {
		tender.ServiceType = models.TenderServiceType(values.ServiceType)
	}

	if values.Status != "" {
		tender.Status = models.TenderStatus(values.Status)
	}
}

func (s *Storage) UpdateTenderByVersion(tender *models.Tender, newTender *models.TenderVersion) {
	tender.Name = newTender.Name
	tender.Description = newTender.Description
	tender.ServiceType = newTender.ServiceType
	tender.Status = newTender.Status
}

func (s *Storage) UpdateTender(tender *models.Tender) error {
	tender.Version++

	if err := s.db.Create(&models.TenderVersion{
		TenderID: tender.ID,
		Version:  tender.Version,
		Name:     tender.Name, Description: tender.Description, ServiceType: tender.ServiceType, Status: tender.Status, EmployeeUsername: tender.EmployeeUsername, OrganizationID: tender.OrganizationID,
	}).Error; err != nil {
		return response.ErrInternalError
	}

	result := s.db.Save(&tender)
	if result.Error != nil {
		return response.ErrInternalError
	}

	return nil
}

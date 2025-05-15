package storage

import (
	"context"

	"github.com/kw3a/spotted-server/internal/database"
	"github.com/kw3a/spotted-server/internal/server/shared"
)

const (
	companyPageSize = 20
)

func (mysql *MysqlStorage) GetCompanyByID(ctx context.Context, companyID string) (shared.Company, error) {
	company, err := mysql.Queries.GetCompanyByID(ctx, companyID)
	if err != nil {
		return shared.Company{}, err
	}
	return shared.Company{
		ID:          company.ID,
		UserID:      company.UserID,
		Name:        company.Name,
		Description: company.Description,
		Website:     company.Website,
		ImageURL:    company.ImageUrl,
	}, nil
}

func (mysql *MysqlStorage) GetCompanies(ctx context.Context, params shared.CompanyQueryParams) ([]shared.Company, error) {
	companies := []shared.Company{}
	var dbCompanies []database.Company
	if params.Query != "" {
		if params.UserID == "" {
			comp, err := mysql.Queries.GetCompaniesByQuery(ctx, database.GetCompaniesByQueryParams{
				Limit:  companyPageSize,
				Offset: (params.Page - 1) * companyPageSize,
				CONCAT: params.Query,
			})
			if err != nil {
				return nil, err
			}
			dbCompanies = comp
		} else {
			comp, err := mysql.Queries.GetCompaniesByUserAndQuery(ctx, database.GetCompaniesByUserAndQueryParams{
				UserID: params.UserID,
				Limit:  companyPageSize,
				Offset: (params.Page - 1) * companyPageSize,
				CONCAT: params.Query,
			})
			if err != nil {
				return nil, err
			}
			dbCompanies = comp
		}
	} else {
		if params.UserID == "" {
			comp, err := mysql.Queries.GetCompanies(ctx, database.GetCompaniesParams{
				Limit:  companyPageSize,
				Offset: (params.Page - 1) * companyPageSize,
			})
			if err != nil {
				return nil, err
			}
			dbCompanies = comp
		} else {
			if params.Page == 0 {
				comp, err := mysql.Queries.GetAllCompaniesByUser(ctx, params.UserID)
				if err != nil {
					return nil, err
				}
				dbCompanies = comp
			} else {
				comp, err := mysql.Queries.GetCompaniesByUser(ctx, database.GetCompaniesByUserParams{
					UserID: params.UserID,
					Limit:  companyPageSize,
					Offset: (params.Page - 1) * companyPageSize,
				})
				if err != nil {
					return nil, err
				}
				dbCompanies = comp
			}
		}
	}
	for _, dbCompany := range dbCompanies {
		companies = append(companies, shared.Company{
			ID:          dbCompany.ID,
			Name:        dbCompany.Name,
			Description: dbCompany.Description,
			Website:     dbCompany.Website,
			ImageURL:    dbCompany.ImageUrl,
		})
	}
	return companies, nil
}

func (mysql *MysqlStorage) GetCompany(ctx context.Context, companyID, userID string) (shared.Company, error) {
	dbCompany, err := mysql.Queries.SelectCompany(ctx, database.SelectCompanyParams{
		ID:     companyID,
		UserID: userID,
	})
	if err != nil {
		return shared.Company{}, err
	}
	return shared.Company{
		ID:          dbCompany.ID,
		Name:        dbCompany.Name,
		Description: dbCompany.Description,
		Website:     dbCompany.Website,
		ImageURL:    dbCompany.ImageUrl,
	}, nil
}

func (mysql *MysqlStorage) RegisterCompany(
	ctx context.Context,
	id string,
	userID string,
	name string,
	description string,
	website string,
	imageURL string,
) error {
	return mysql.Queries.InsertCompany(ctx, database.InsertCompanyParams{
		ID:          id,
		UserID:      userID,
		Name:        name,
		Description: description,
		Website:     website,
		ImageUrl:    imageURL,
	})
}

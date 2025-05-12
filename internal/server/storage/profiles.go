package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/kw3a/spotted-server/internal/database"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"golang.org/x/crypto/bcrypt"
)

func (mysql *MysqlStorage) UpdateDescription(ctx context.Context, userID, description string) error {
	return mysql.Queries.UpdateUserDescription(ctx, database.UpdateUserDescriptionParams{
		ID:          userID,
		Description: description,
	})
}

func (mysql *MysqlStorage) UpdateProfilePic(ctx context.Context, userID string, imageURL string) error {
	return mysql.Queries.UpdateImage(ctx, database.UpdateImageParams{
		ID:       userID,
		ImageUrl: imageURL,
	})
}

func (mysql *MysqlStorage) DeleteEducation(ctx context.Context, userID string, educationID string) error {
	return mysql.Queries.DeleteEducation(ctx, database.DeleteEducationParams{
		ID:     educationID,
		UserID: userID,
	})
}

func (mysql *MysqlStorage) RegisterEducation(
	ctx context.Context,
	educationID, userID, institution, degree string,
	start time.Time,
	end sql.NullTime,
) error {
	return mysql.Queries.InsertEducation(ctx, database.InsertEducationParams{
		ID:          educationID,
		Institution: institution,
		Degree:      degree,
		StartDate:   start,
		EndDate:     end,
		UserID:      userID,
	})
}

func (mysql *MysqlStorage) DeleteExperience(ctx context.Context, userID string, experienceID string) error {
	return mysql.Queries.DeleteExperience(ctx, database.DeleteExperienceParams{
		ID:     experienceID,
		UserID: userID,
	})
}

func (mysql *MysqlStorage) RegisterExperience(
	ctx context.Context,
	experienceID, userID, company, title string,
	start time.Time,
	end sql.NullTime,
) error {
	return mysql.Queries.InsertExperience(ctx, database.InsertExperienceParams{
		ID:        experienceID,
		Company:   company,
		Title:     title,
		StartDate: start,
		EndDate:   end,
		UserID:    userID,
	})
}

func (mysql *MysqlStorage) DeleteSkill(ctx context.Context, userID string, skillID string) error {
	return mysql.Queries.DeleteSkill(ctx, database.DeleteSkillParams{
		ID:     skillID,
		UserID: userID,
	})
}

func (mysql *MysqlStorage) RegisterSkill(ctx context.Context, skillID, userID, name string) error {
	return mysql.Queries.InsertSkill(ctx, database.InsertSkillParams{
		ID:     skillID,
		Name:   name,
		UserID: userID,
	})
}

func (mysql *MysqlStorage) DeleteLink(ctx context.Context, userID string, linkID string) error {
	return mysql.Queries.DeleteLink(ctx, database.DeleteLinkParams{
		ID:     linkID,
		UserID: userID,
	})
}

func (mysql *MysqlStorage) RegisterLink(ctx context.Context, linkID, userID, url, name string) error {
	return mysql.Queries.InsertLink(ctx, database.InsertLinkParams{
		ID:     linkID,
		Url:    url,
		Name:   name,
		UserID: userID,
	})
}

func (mysql *MysqlStorage) SelectEducation(ctx context.Context, userID string) ([]shared.EducationEntry, error) {
	education, err := mysql.Queries.SelectEducation(ctx, userID)
	res := []shared.EducationEntry{}
	if err == sql.ErrNoRows {
		return res, nil
	}
	if err != nil {
		return nil, err
	}
	for _, entry := range education {
		startDate := shared.DateSpanishFormat(sql.NullTime{Valid: true, Time: entry.StartDate})

		endDate := shared.DateSpanishFormat(entry.EndDate)

		res = append(res, shared.EducationEntry{
			ID:          entry.ID,
			Degree:      entry.Degree,
			Institution: entry.Institution,
			StartDate:   startDate,
			EndDate:     endDate,
		})
	}
	return res, nil
}

func (mysql *MysqlStorage) SelectExperiences(ctx context.Context, userID string) ([]shared.ExperienceEntry, error) {
	experiences, err := mysql.Queries.SelectExperience(ctx, userID)
	res := []shared.ExperienceEntry{}
	if err == sql.ErrNoRows {
		return res, nil
	}
	if err != nil {
		return nil, err
	}
	for _, entry := range experiences {
		startDate := shared.DateSpanishFormat(sql.NullTime{Valid: true, Time: entry.StartDate})
		endDate := shared.DateSpanishFormat(entry.EndDate)
		res = append(res, shared.ExperienceEntry{
			ID:           entry.ID,
			Title:        entry.Title,
			Company:      entry.Company,
			StartDate:    startDate,
			EndDate:      endDate,
			TimeInterval: shared.TimeInterval(entry.StartDate, entry.EndDate),
		})
	}
	return res, nil
}

func (mysql *MysqlStorage) SelectLinks(ctx context.Context, userID string) ([]shared.Link, error) {
	res := []shared.Link{}
	links, err := mysql.Queries.SelectLinks(ctx, userID)
	if err == sql.ErrNoRows {
		return res, nil
	}
	if err != nil {
		return nil, err
	}
	for _, link := range links {
		res = append(res, shared.Link{
			URL:  link.Url,
			Name: link.Name,
			ID:   link.ID,
		})
	}
	return res, nil
}

func (mysql *MysqlStorage) SelectSkills(ctx context.Context, userID string) ([]shared.SkillEntry, error) {
	skills, err := mysql.Queries.SelectSkills(ctx, userID)
	res := []shared.SkillEntry{}
	if err == sql.ErrNoRows {
		return res, nil
	}
	if err != nil {
		return nil, err
	}
	for _, skill := range skills {
		res = append(res, shared.SkillEntry{
			Name: skill.Name,
			ID:   skill.ID,
		})
	}
	return res, nil
}


func (mysql *MysqlStorage) CreateUser(ctx context.Context, id, nick, name, password string) error {
	encriptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return mysql.Queries.CreateUser(ctx, database.CreateUserParams{
		ID:       id,
		Nick:     nick,
		Name:     name,
		Password: string(encriptedPassword),
	})
}

func (mysql *MysqlStorage) UpdateCell(ctx context.Context, userID string, cell string) error {
	return mysql.Queries.UpdateCell(ctx, database.UpdateCellParams{ID: userID, Number: cell})
}

func (mysql *MysqlStorage) UpdateEmail(ctx context.Context, userID string, email string) error {
	return mysql.Queries.UpdateEmail(ctx, database.UpdateEmailParams{ID: userID, Email: email})
}

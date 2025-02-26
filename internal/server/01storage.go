package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/database"
	"github.com/kw3a/spotted-server/internal/server/codejudge"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"golang.org/x/crypto/bcrypt"
)

type MysqlStorage struct {
	Queries *database.Queries
	db      *sql.DB
}

func (mysql *MysqlStorage) RegisterOffer(
	ctx context.Context,
	offerID string,
	offer shared.Offer,
	quizID string,
	quiz shared.Quiz,
	problems []shared.Problem,
) error {
	tx, err := mysql.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := mysql.Queries.WithTx(tx)
	err = qtx.InsertOffer(ctx, database.InsertOfferParams{
		ID:           offerID,
		Title:        offer.Title,
		About:        offer.About,
		Requirements: offer.Requirements,
		Benefits:     offer.Benefits,
		MinWage:      offer.MinWage,
		MaxWage:      offer.MaxWage,
		CompanyID:    offer.CompanyID,
	})

	if err != nil {
		return fmt.Errorf("error inserting offer: %w", err)
	}
	err = qtx.InsertQuiz(ctx, database.InsertQuizParams{
		ID:       quizID,
		Duration: quiz.Duration,
		OfferID:  offerID,
	})
	if err != nil {
		return fmt.Errorf("error inserting quiz: %w", err)
	}
	for _, lang := range quiz.Languages {
		err = qtx.InsertLanguageQuiz(ctx, database.InsertLanguageQuizParams{
			ID:         uuid.New().String(),
			QuizID:     quizID,
			LanguageID: lang,
		})
		if err != nil {
			return fmt.Errorf("error inserting language quiz: %w", err)
		}
	}
	for _, problem := range problems {
		problemID := uuid.New().String()
		err = qtx.InsertProblem(ctx, database.InsertProblemParams{
			ID:          problemID,
			QuizID:      quizID,
			Title:       problem.Title,
			Description: problem.Description,
			TimeLimit:   float64(problem.TimeLimit),
			MemoryLimit: 262144,
		})
		if err != nil {
			return fmt.Errorf("error inserting problem: %w", err)
		}
		for _, tc := range problem.TestCases {
			err = qtx.InsertTestCase(ctx, database.InsertTestCaseParams{
				ID:        uuid.New().String(),
				ProblemID: problemID,
				Input:     tc.Input,
				Output:    tc.Output,
			})
			if err != nil {
				return fmt.Errorf("error inserting test case: %w", err)
			}
		}
		for _, example := range problem.Examples {
			err = qtx.InsertExample(ctx, database.InsertExampleParams{
				ID:        uuid.New().String(),
				ProblemID: problemID,
				Input:     example.Input,
				Output:    example.Output,
			})
			if err != nil {
				return fmt.Errorf("error inserting example: %w", err)
			}
		}
	}
	return tx.Commit()
}

func (mysql *MysqlStorage) InsertQuiz(r *http.Request, quizID string, offerID string, languages []int32, duration int32) error {
	tx, err := mysql.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := mysql.Queries.WithTx(tx)
	err = qtx.InsertQuiz(r.Context(), database.InsertQuizParams{
		ID:       quizID,
		OfferID:  offerID,
		Duration: duration,
	})
	if err != nil {
		fmt.Println(offerID)
		return err
	}
	for _, lang := range languages {
		err = qtx.InsertLanguageQuiz(r.Context(), database.InsertLanguageQuizParams{
			ID:         uuid.New().String(),
			QuizID:     quizID,
			LanguageID: lang,
		})
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (mysql *MysqlStorage) SelectQuizByOffer(ctx context.Context, offerID string) (shared.Quiz, error) {
	dbQuiz, err := mysql.Queries.GetQuizByOffer(ctx, offerID)
	if err != nil {
		return shared.Quiz{}, err
	}
	return shared.Quiz{
		ID:       dbQuiz.ID,
		Duration: dbQuiz.Duration,
	}, nil
}

func (mysql *MysqlStorage) SelectOfferByUser(ctx context.Context, ID string, userID string) (shared.Offer, error) {
	dbQuiz, err := mysql.Queries.GetOfferByUser(ctx, database.GetOfferByUserParams{
		ID:   ID,
		ID_2: userID,
	})
	if err != nil {
		return shared.Offer{}, err
	}
	relativeTime := RelativeTime(dbQuiz.CreatedAt)
	return shared.Offer{
		ID:           dbQuiz.ID,
		Title:        dbQuiz.Title,
		About:        dbQuiz.About,
		Status:       dbQuiz.Status,
		CompanyName:  dbQuiz.CompanyName,
		CompanyID:    dbQuiz.CompanyID,
		MinWage:      dbQuiz.MinWage,
		MaxWage:      dbQuiz.MaxWage,
		RelativeTime: relativeTime,
	}, nil
}

func (mysql *MysqlStorage) GetOffersByUser(ctx context.Context, userID string) ([]shared.Offer, error) {
	offers, err := mysql.Queries.GetOffersByUser(ctx, userID)
	if err == sql.ErrNoRows {
		return []shared.Offer{}, nil
	}
	if err != nil {
		return nil, err
	}
	res := []shared.Offer{}
	for _, offer := range offers {
		relativeTime := RelativeTime(offer.CreatedAt)
		res = append(res, shared.Offer{
			ID:           offer.ID,
			Title:        offer.Title,
			About:        offer.About,
			Requirements: offer.Requirements,
			Benefits:     offer.Benefits,
			Status:       offer.Status,
			CompanyName:  offer.CompanyName,
			CompanyID:    offer.CompanyID,
			MinWage:      offer.MinWage,
			MaxWage:      offer.MaxWage,
			RelativeTime: relativeTime,
		})
	}
	return res, nil

}

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
		Website:     company.Website.String,
		ImageURL:    company.ImageUrl.String,
	}, nil
}

func (mysql *MysqlStorage) GetLanguages(ctx context.Context) ([]shared.Language, error) {
	languages, err := mysql.Queries.AllLanguages(ctx)
	if err != nil {
		return nil, err
	}
	res := []shared.Language{}
	for _, lang := range languages {
		res = append(res, shared.Language{
			ID:          lang.ID,
			Name:        lang.Name,
			DisplayName: lang.DisplayName,
		})
	}
	return res, nil
}

func (mysql *MysqlStorage) GetCompanies(ctx context.Context, params shared.CompanyQueryParams) ([]shared.Company, error) {
	companies := []shared.Company{}
	var dbCompanies []database.Company
	if params.Query != "" {
		if params.UserID == "" {
			comp, err := mysql.Queries.GetCompaniesByQuery(ctx, params.Query)
			if err != nil {
				return nil, err
			}
			dbCompanies = comp
		} else {
			comp, err := mysql.Queries.GetCompaniesByUserAndQuery(ctx, database.GetCompaniesByUserAndQueryParams{
				UserID: params.UserID,
				CONCAT: params.Query,
			})
			if err != nil {
				return nil, err
			}
			dbCompanies = comp
		}
	} else {
		if params.UserID == "" {
			comp, err := mysql.Queries.GetCompanies(ctx)
			if err != nil {
				return nil, err
			}
			dbCompanies = comp
		} else {
			comp, err := mysql.Queries.GetCompaniesByUser(ctx, params.UserID)
			if err != nil {
				return nil, err
			}
			dbCompanies = comp
		}
	}
	for _, dbCompany := range dbCompanies {
		companies = append(companies, shared.Company{
			ID:          dbCompany.ID,
			Name:        dbCompany.Name,
			Description: dbCompany.Description,
			Website:     dbCompany.Website.String,
			ImageURL:    dbCompany.ImageUrl.String,
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
		Website:     dbCompany.Website.String,
		ImageURL:    dbCompany.ImageUrl.String,
	}, nil
}

func (mysql *MysqlStorage) RegisterCompany(ctx context.Context, id string, userID string, name string, description string, website string, imageURL string) error {
	nullWebsite := sql.NullString{String: website, Valid: true}
	nullImageURL := sql.NullString{String: imageURL, Valid: true}
	return mysql.Queries.InsertCompany(ctx, database.InsertCompanyParams{
		ID:          id,
		UserID:      userID,
		Name:        name,
		Description: description,
		Website:     nullWebsite,
		ImageUrl:    nullImageURL,
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

func (mysql *MysqlStorage) RegisterEducation(ctx context.Context, educationID, userID, institution, degree string, start time.Time, end time.Time) error {
	return mysql.Queries.InsertEducation(ctx, database.InsertEducationParams{
		ID:          educationID,
		Institution: institution,
		Degree:      degree,
		StartDate:   start,
		EndDate:     sql.NullTime{Time: end, Valid: true},
		UserID:      userID,
	})
}

func (mysql *MysqlStorage) DeleteExperience(ctx context.Context, userID string, experienceID string) error {
	return mysql.Queries.DeleteExperience(ctx, database.DeleteExperienceParams{
		ID:     experienceID,
		UserID: userID,
	})
}

func (mysql *MysqlStorage) RegisterExperience(ctx context.Context, experienceID, userID, company, title string, start time.Time, end time.Time) error {
	return mysql.Queries.InsertExperience(ctx, database.InsertExperienceParams{
		ID:        experienceID,
		Company:   company,
		Title:     title,
		StartDate: start,
		EndDate:   sql.NullTime{Time: end, Valid: true},
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
		endDate := "actualmente"
		if entry.EndDate.Valid {
			endDate = entry.EndDate.Time.Format("2006-01")
		}
		res = append(res, shared.EducationEntry{
			ID:          entry.ID,
			Degree:      entry.Degree,
			Institution: entry.Institution,
			StartDate:   entry.StartDate.Format("2006-01"),
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
		endDate := "actualmente"
		if entry.EndDate.Valid {
			endDate = entry.EndDate.Time.Format("2006-01")
		}
		res = append(res, shared.ExperienceEntry{
			ID:        entry.ID,
			Title:     entry.Title,
			Company:   entry.Company,
			StartDate: entry.StartDate.Format("2006-01"),
			EndDate:   endDate,
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

func NewMysqlStorage(dbURL string) (*MysqlStorage, error) {
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	queries := database.New(db)

	return &MysqlStorage{Queries: queries, db: db}, nil
}

func (mysql *MysqlStorage) CreateUser(ctx context.Context, name, password, email, description string) error {
	encriptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return mysql.Queries.CreateUser(ctx, database.CreateUserParams{
		ID:          uuid.New().String(),
		Name:        name,
		Email:       email,
		Password:    string(encriptedPassword),
		Role:        "dev",
		Description: description,
	})
}

func (mysql *MysqlStorage) ParticipationStatus(ctx context.Context, userID string, quizID string) (ParticipationData, error) {
	participation, err := mysql.Queries.ParticipationStatus(ctx, database.ParticipationStatusParams{
		UserID: userID,
		QuizID: quizID,
	})
	if err != nil {
		return ParticipationData{}, err
	}
	relativeTime := RelativeTime(participation.ExpiresAt)
	return ParticipationData{
		ID:           participation.ID,
		CreatedAt:    participation.CreatedAt.Time,
		ExpiresAt:    participation.ExpiresAt,
		RelativeTime: relativeTime,
	}, nil
}

func (mysql *MysqlStorage) Participate(ctx context.Context, userID string, quizID string) error {
	return mysql.Queries.Participate(ctx, database.ParticipateParams{
		ID:     uuid.New().String(),
		UserID: userID,
		QuizID: quizID,
	})
}

func (mysql *MysqlStorage) SelectProblemIDs(ctx context.Context, QuizID string) ([]string, error) {
	IDs, err := mysql.Queries.SelectProblemIDs(ctx, QuizID)
	if err != nil {
		return nil, err
	}
	if len(IDs) == 0 {
		return nil, errors.New("zero results")
	}
	return IDs, nil
}

func (mysql *MysqlStorage) SelectOffer(ctx context.Context, ID string) (shared.Offer, error) {
	dbQuiz, err := mysql.Queries.GetOffer(ctx, ID)
	if err != nil {
		return shared.Offer{}, err
	}
	relativeTime := RelativeTime(dbQuiz.CreatedAt)
	return shared.Offer{
		ID:           dbQuiz.ID,
		Title:        dbQuiz.Title,
		About:        dbQuiz.About,
		Requirements: dbQuiz.Requirements,
		Benefits:     dbQuiz.Benefits,
		Status:       dbQuiz.Status,
		CompanyName:  dbQuiz.CompanyName,
		CompanyID:    dbQuiz.CompanyID,
		MinWage:      dbQuiz.MinWage,
		MaxWage:      dbQuiz.MaxWage,
		RelativeTime: relativeTime,
	}, nil
}

func ConvertLanguages(languages sql.NullString) ([]string, error) {
	if !languages.Valid {
		return nil, errors.New("languages is NULL")
	}
	result := strings.Split(languages.String, ",")
	if len(result) == 0 {
		return nil, errors.New("no languages found")
	}
	return result, nil
}

func (mysql *MysqlStorage) SelectOffers(ctx context.Context, params shared.JobQueryParams) ([]shared.Offer, error) {
	offers := []shared.Offer{}
	if params.Query == "" {
		quizzes, err := mysql.Queries.GetOffers(ctx)
		if err != nil {
			return nil, err
		}
		for _, quiz := range quizzes {
			relativeTime := RelativeTime(quiz.CreatedAt)
			offers = append(offers, shared.Offer{
				ID:              quiz.ID,
				Title:           quiz.Title,
				About:           quiz.About,
				Requirements:    quiz.Requirements,
				Benefits:        quiz.Benefits,
				Status:          quiz.Status,
				CompanyName:     quiz.CompanyName,
				CompanyID:       quiz.CompanyID,
				CompanyImageURL: quiz.CompanyImageUrl.String,
				MinWage:         quiz.MinWage,
				MaxWage:         quiz.MaxWage,
				RelativeTime:    relativeTime,
			})
		}
	} else {
		quizzes, err := mysql.Queries.GetOffersByQuery(ctx, params.Query)
		if err != nil {
			return nil, err
		}
		for _, quiz := range quizzes {
			relativeTime := RelativeTime(quiz.CreatedAt)
			offers = append(offers, shared.Offer{
				ID:              quiz.ID,
				Title:           quiz.Title,
				About:           quiz.About,
				Requirements:    quiz.Requirements,
				Benefits:        quiz.Benefits,
				Status:          quiz.Status,
				CompanyName:     quiz.CompanyName,
				CompanyID:       quiz.CompanyID,
				CompanyImageURL: quiz.CompanyImageUrl.String,
				MinWage:         quiz.MinWage,
				MaxWage:         quiz.MaxWage,
				RelativeTime:    relativeTime,
			})
		}
	}
	return offers, nil
}

func RelativeTime(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	if duration < 0 {
		duration = -duration
		if duration < time.Minute {
			return "in a few seconds"
		} else if duration < time.Hour {
			return fmt.Sprintf("in %d minutes", int(duration.Minutes()))
		} else if duration < 24*time.Hour {
			return fmt.Sprintf("in %d hours", int(duration.Hours()))
		} else {
			days := int(duration.Hours() / 24)
			return fmt.Sprintf("in %d days", days)
		}
	}
	if duration < time.Minute {
		return "justo ahora"
	} else if duration < time.Hour {
		return fmt.Sprintf("hace %d minutos", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("hace %d horas", int(duration.Hours()))
	} else {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("hace %d días", days)
	}
}

func (mysql *MysqlStorage) SelectLanguages(ctx context.Context, quizID string) ([]shared.Language, error) {
	dbLanguages, err := mysql.Queries.SelectLanguages(ctx, quizID)
	if err != nil {
		return nil, err
	}
	res := []shared.Language{}
	for _, dbLang := range dbLanguages {
		lang := shared.Language{
			ID:          dbLang.ID,
			Name:        dbLang.Name,
			DisplayName: dbLang.DisplayName,
		}
		res = append(res, lang)
	}
	return res, nil
}

func (mysql *MysqlStorage) SelectScore(ctx context.Context, userID string, problemID string) (ScoreData, error) {
	totalTestCases, err := mysql.Queries.TotalTestCases(ctx, problemID)
	if err != nil {
		return ScoreData{}, err
	}
	best, err := mysql.Queries.BestSubmission(ctx, database.BestSubmissionParams{
		ProblemID: problemID,
		UserID:    userID,
	})
	acceptedTestCases := int(best.AcceptedTestCases)
	if err == sql.ErrNoRows {
		acceptedTestCases = 0
	}
	return ScoreData{
		AcceptedTestCases: acceptedTestCases,
		TotalTestCases:    int(totalTestCases),
	}, nil
}

func (mysql *MysqlStorage) SelectProblem(ctx context.Context, problemID string) (ProblemContent, error) {
	dbProblem, err := mysql.Queries.SelectProblem(ctx, problemID)
	if err != nil {
		return ProblemContent{}, err
	}
	return ProblemContent{
		Title:       dbProblem.Title,
		Description: dbProblem.Description,
		MemoryLimit: dbProblem.MemoryLimit,
		TimeLimit:   dbProblem.TimeLimit,
	}, nil
}

func (mysql *MysqlStorage) SelectExamples(ctx context.Context, problemID string) ([]Example, error) {
	dbExamples, err := mysql.Queries.SelectExamples(ctx, problemID)
	if err != nil {
		return []Example{}, err
	}
	res := []Example{}
	for _, dbExample := range dbExamples {
		res = append(res, Example{
			Input:  dbExample.Input,
			Output: dbExample.Output,
		})
	}
	return res, nil
}

func (mysql *MysqlStorage) LastSrc(ctx context.Context, userID string, problemID string, languageID int32) (string, error) {
	dbLastSubmission, err := mysql.Queries.LastSubmission(ctx, database.LastSubmissionParams{
		ProblemID:  problemID,
		LanguageID: languageID,
		UserID:     userID,
	})
	if err == sql.ErrNoRows {
		return ExampleCode(languageID)
	}
	if err != nil {
		return "", err
	}
	return dbLastSubmission, nil
}

func (s MysqlStorage) CreateSubmission(ctx context.Context, submissionID, participationID, problemID, src string, languageID int32) error {
	return s.Queries.CreateSubmission(ctx, database.CreateSubmissionParams{
		ID:              submissionID,
		Src:             src,
		LanguageID:      languageID,
		ProblemID:       problemID,
		ParticipationID: participationID,
	})
}

func (s MysqlStorage) GetTestCases(ctx context.Context, problemID string) ([]codejudge.TestCase, error) {
	dbTestCases, err := s.Queries.GetTestCases(ctx, problemID)
	if err != nil {
		return nil, err
	}
	return ToTC(dbTestCases)
}

func ToTC(dbTestCases []database.GetTestCasesRow) ([]codejudge.TestCase, error) {
	if len(dbTestCases) == 0 {
		return []codejudge.TestCase{}, errors.New("empty database test cases")
	}
	res := []codejudge.TestCase{}
	for _, testCase := range dbTestCases {
		current := codejudge.TestCase{
			ID:             testCase.ID,
			TimeLimit:      testCase.TimeLimit,
			MemoryLimit:    testCase.MemoryLimit,
			Input:          testCase.Input,
			ExpectedOutput: testCase.Output,
		}
		res = append(res, current)
	}
	return res, nil
}

// This method must work only when TCResult is empty
func (s MysqlStorage) UpdateTestCaseResult(ctx context.Context, input CallbackJsonInput, submissionID, testCaseID string) error {
	return s.Queries.UpdateTestCaseResult(ctx, database.UpdateTestCaseResultParams{
		ID:           sql.NullString{String: input.Token, Valid: true},
		Status:       sql.NullString{String: input.Status.Description, Valid: true},
		Time:         sql.NullString{String: input.Time.String(), Valid: true},
		Memory:       sql.NullInt32{Int32: input.Memory, Valid: true},
		SubmissionID: submissionID,
		TestCaseID:   testCaseID,
	})
}
func (s MysqlStorage) EndQuiz(ctx context.Context, userID, quizID string) (shared.Offer, error) {
	err := s.Queries.EndParticipation(ctx, database.EndParticipationParams{
		ExpiresAt: time.Now(),
		UserID:    userID,
		QuizID:    quizID,
	})
	if err != nil {
		return shared.Offer{}, err
	}
	offer, _ := s.Queries.GetOfferByQuiz(ctx, quizID)
	return shared.Offer{
		ID: offer.ID,
	}, nil
}

func CheckPasswordHash(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
func (s *MysqlStorage) GetUserID(ctx context.Context, email, password string) (string, error) {
	dbUser, err := s.Queries.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("email and password do not match")
	}
	err = CheckPasswordHash(dbUser.Password, password)
	if err != nil {
		return "", errors.New("email and password do not match")
	}
	return dbUser.ID, nil
}
func (s *MysqlStorage) IsRegistered(ctx context.Context, refreshToken string) error {
	tokenExists, err := s.Queries.VerifyRefreshJWT(ctx, refreshToken)
	if err != nil {
		return err
	}
	if !tokenExists {
		return errors.New("token does not exist")
	}
	return nil
}
func (s *MysqlStorage) Save(ctx context.Context, refreshToken string) error {
	return s.Queries.SaveRefreshJWT(ctx, database.SaveRefreshJWTParams{
		RefreshToken: refreshToken,
		CreatedAt:    time.Now().UTC(),
	})
}
func (s *MysqlStorage) Revoke(ctx context.Context, refreshToken string) error {
	err := s.IsRegistered(ctx, refreshToken)
	if err != nil {
		return err
	}
	return s.Queries.RevokeRefreshJWT(ctx, refreshToken)
}

func (s *MysqlStorage) GetUser(ctx context.Context, userID string) (database.User, error) {
	return s.Queries.GetUser(ctx, userID)
}

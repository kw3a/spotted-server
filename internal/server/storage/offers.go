package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/internal/database"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"github.com/shopspring/decimal"
)

const (
	offerPageSize = 20
)

func (mysql *MysqlStorage) SelectFullProblems(ctx context.Context, quizID string) ([]shared.Problem, error) {
	// Step 1: Get all problems
	dbProblems, err := mysql.Queries.SelectProblems(ctx, quizID)
	if err != nil {
		return nil, err
	}

	// Step 2: Extract problem IDs
	problemIDs := make([]string, len(dbProblems))
	for i, p := range dbProblems {
		problemIDs[i] = p.ID
	}

	// Step 3: Fetch test cases
	dbTestCases, err := mysql.Queries.BatchTestCases(ctx, problemIDs)
	if err != nil {
		return nil, err
	}

	// Step 4: Fetch examples
	dbExamples, err := mysql.Queries.BatchExamples(ctx, problemIDs)
	if err != nil {
		return nil, err
	}

	// Step 5: Group test cases by problem ID
	testCaseMap := make(map[string][]shared.TestCase)
	for _, tc := range dbTestCases {
		tcStruct := shared.TestCase{
			ID:        tc.ID,
			Input:     tc.Input,
			Output:    tc.Output,
			ProblemID: tc.ProblemID,
		}
		testCaseMap[tc.ProblemID] = append(testCaseMap[tc.ProblemID], tcStruct)
	}

	// Step 6: Group examples by problem ID
	exampleMap := make(map[string][]shared.Example)
	for _, ex := range dbExamples {
		exStruct := shared.Example{
			ID:        ex.ID,
			Input:     ex.Input,
			Output:    ex.Output,
			ProblemID: ex.ProblemID,
		}
		exampleMap[ex.ProblemID] = append(exampleMap[ex.ProblemID], exStruct)
	}

	// Step 7: Build final problem slice
	res := make([]shared.Problem, 0, len(dbProblems))
	for _, p := range dbProblems {
		prob := shared.Problem{
			ID:          p.ID,
			Title:       p.Title,
			Description: p.Description,
			MemoryLimit: p.MemoryLimit,
			TimeLimit:   p.TimeLimit,
			TestCases:   testCaseMap[p.ID],
			Examples:    exampleMap[p.ID],
		}
		res = append(res, prob)
	}

	return res, nil
}

func (mysql *MysqlStorage) SelectApplications(ctx context.Context, quizID string) ([]shared.Application, error) {
	applications, err := mysql.selectApplicants(ctx, quizID)
	if err != nil {
		return nil, err
	}

	participationIDs := make([]string, 0, len(applications))
	applicationIndex := make(map[string]*shared.Application)
	for i := range applications {
		participationID := applications[i].Participation.ID
		participationIDs = append(participationIDs, participationID)
		applicationIndex[participationID] = &applications[i]
	}

	summaries, err := mysql.batchSummary(ctx, quizID, participationIDs)
	if err != nil {
		return nil, err
	}

	summariesByParticipation := make(map[string][]shared.Summary)
	submissionIDs := []string{}
	for _, summary := range summaries {
		pid := summary.Submission.ParticipationID
		summariesByParticipation[pid] = append(summariesByParticipation[pid], summary)
		submissionIDs = append(submissionIDs, summary.Submission.ID)
	}

	testCaseResults, err := mysql.batchTCResults(ctx, submissionIDs)
	if err != nil {
		return nil, err
	}

	resultsBySubmission := make(map[string][]shared.TestCaseResult)
	for _, result := range testCaseResults {
		resultsBySubmission[result.SubmissionID] = append(resultsBySubmission[result.SubmissionID], result)
	}

	finalApplications := []shared.Application{}
	for _, application := range applications {
		participationID := application.Participation.ID
		summaries := summariesByParticipation[participationID]
		for i := range summaries {
			summaries[i].Results = resultsBySubmission[summaries[i].Submission.ID]
		}
		application.Summary = summaries
		finalApplications = append(finalApplications, application)
	}

	return finalApplications, nil
}

func (mysql *MysqlStorage) selectApplicants(ctx context.Context, quizID string) ([]shared.Application, error) {
	applications, err := mysql.Queries.SelectApplications(ctx, quizID)
	if err != nil {
		return []shared.Application{}, err
	}
	res := []shared.Application{}
	for _, apl := range applications {
		ap := shared.Application{
			Applicant: shared.User{
				ID:          apl.ID,
				Name:        apl.Name,
				ImageURL:    apl.ImageUrl,
				Email:       apl.Email,
				Description: apl.Description,
			},
			Participation: shared.Participation{
				ID:           apl.ParticipationID,
				CreatedAt:    apl.ParticipationCreatedAt.Time,
				ExpiresAt:    apl.ParticipationExpiresAt,
				RelativeTime: RelativeTime(apl.ParticipationExpiresAt),
			},
		}
		res = append(res, ap)
	}
	return res, nil
}

func (mysql *MysqlStorage) batchSummary(
	ctx context.Context,
	quizID string,
	participationIDs []string,
) ([]shared.Summary, error) {
	sum, err := mysql.Queries.BatchBestSubmissionsFromParticipation(ctx, participationIDs)
	if err != nil {
		return []shared.Summary{}, err
	}
	res := []shared.Summary{}
	for _, s := range sum {
		current := shared.Summary{
			Title: s.Title,
			Score: shared.Score{
				AcceptedTestCases: int(s.AcceptedTestCases),
				TotalTestCases:    int(s.TotalTestCases),
			},
			Submission: shared.Submission{
				ID:              s.ID,
				Src:             s.Src,
				LanguageID:      s.LanguageID,
				Language:        s.DisplayName,
				ProblemID:       s.ProblemID,
				ParticipationID: s.ParticipationID,
			},
		}
		res = append(res, current)
	}
	fmt.Println("summary length: ", len(res))
	return res, nil
}

func (mysql *MysqlStorage) batchTCResults(
	ctx context.Context,
	submissionIDs []string,
) ([]shared.TestCaseResult, error) {
	dbTCResults, err := mysql.Queries.BatchTCResults(ctx, submissionIDs)
	if err != nil {
		return []shared.TestCaseResult{}, err
	}
	res := []shared.TestCaseResult{}
	for _, dbTCResult := range dbTCResults {
		dbTime, err := decimal.NewFromString(dbTCResult.Time)
		if err != nil {
			return []shared.TestCaseResult{}, err
		}
		dbTime = dbTime.Mul(decimal.NewFromInt32(1000))
		current := shared.TestCaseResult{
			ID:           dbTCResult.ID.String,
			Output:       dbTCResult.Output,
			Status:       dbTCResult.Status,
			Time:         dbTime.IntPart(),
			Memory:       dbTCResult.Memory,
			TestCaseID:   dbTCResult.TestCaseID,
			SubmissionID: dbTCResult.SubmissionID,
		}
		res = append(res, current)
	}
	fmt.Println("tc length: ", len(res))
	return res, nil
}

func (mysql *MysqlStorage) ArchiveOffer(ctx context.Context, offerID string, ownerID string) error {
	return mysql.Queries.ArchiveOffer(ctx, database.ArchiveOfferParams{
		ID:     offerID,
		UserID: ownerID,
	})
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
		ID:              dbQuiz.ID,
		Title:           dbQuiz.Title,
		About:           dbQuiz.About,
		Status:          dbQuiz.Status,
		CompanyName:     dbQuiz.CompanyName,
		CompanyID:       dbQuiz.CompanyID,
		CompanyImageURL: dbQuiz.CompanyImageUrl,
		MinWage:         dbQuiz.MinWage,
		MaxWage:         dbQuiz.MaxWage,
		RelativeTime:    relativeTime,
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

func (mysql *MysqlStorage) SelectOffer(ctx context.Context, ID string) (shared.Offer, error) {
	dbQuiz, err := mysql.Queries.GetOffer(ctx, ID)
	if err != nil {
		return shared.Offer{}, err
	}
	relativeTime := RelativeTime(dbQuiz.CreatedAt)
	return shared.Offer{
		ID:              dbQuiz.ID,
		Title:           dbQuiz.Title,
		About:           dbQuiz.About,
		Requirements:    dbQuiz.Requirements,
		Benefits:        dbQuiz.Benefits,
		Status:          dbQuiz.Status,
		CompanyName:     dbQuiz.CompanyName,
		CompanyID:       dbQuiz.CompanyID,
		CompanyImageURL: dbQuiz.CompanyImageUrl,
		MinWage:         dbQuiz.MinWage,
		MaxWage:         dbQuiz.MaxWage,
		RelativeTime:    relativeTime,
	}, nil
}

func (mysql *MysqlStorage) GetOffersByCompany(ctx context.Context, companyID string, page int32) ([]shared.Offer, error) {
	offers := []shared.Offer{}
	dbOffers, err := mysql.Queries.GetOffersByCompany(ctx, database.GetOffersByCompanyParams{
		ID:     companyID,
		Limit:  offerPageSize,
		Offset: (page - 1) * offerPageSize,
	})
	if err != nil {
		return nil, err
	}
	for _, dbOffer := range dbOffers {
		relativeTime := RelativeTime(dbOffer.CreatedAt)
		offers = append(offers, shared.Offer{
			ID:              dbOffer.ID,
			Title:           dbOffer.Title,
			About:           dbOffer.About,
			Requirements:    dbOffer.Requirements,
			Benefits:        dbOffer.Benefits,
			Status:          dbOffer.Status,
			CompanyName:     dbOffer.CompanyName,
			CompanyID:       dbOffer.CompanyID,
			CompanyImageURL: dbOffer.CompanyImageUrl,
			MinWage:         dbOffer.MinWage,
			MaxWage:         dbOffer.MaxWage,
			RelativeTime:    relativeTime,
		})
	}
	return offers, nil
}

func (mysql *MysqlStorage) GetOffers(ctx context.Context, page int32) ([]shared.Offer, error) {
	offers := []shared.Offer{}
	dbOffers, err := mysql.Queries.GetOffers(ctx, database.GetOffersParams{
		Limit:  offerPageSize,
		Offset: (page - 1) * offerPageSize,
	})
	if err != nil {
		return nil, err
	}
	for _, dbOffer := range dbOffers {
		relativeTime := RelativeTime(dbOffer.CreatedAt)
		offers = append(offers, shared.Offer{
			ID:              dbOffer.ID,
			Title:           dbOffer.Title,
			About:           dbOffer.About,
			Requirements:    dbOffer.Requirements,
			Benefits:        dbOffer.Benefits,
			Status:          dbOffer.Status,
			CompanyName:     dbOffer.CompanyName,
			CompanyID:       dbOffer.CompanyID,
			CompanyImageURL: dbOffer.CompanyImageUrl,
			MinWage:         dbOffer.MinWage,
			MaxWage:         dbOffer.MaxWage,
			RelativeTime:    relativeTime,
		})
	}
	return offers, nil
}

func (mysql *MysqlStorage) GetOffersByQuery(ctx context.Context, query string, page int32) ([]shared.Offer, error) {
	offers := []shared.Offer{}
	dbOffers, err := mysql.Queries.GetOffersByQuery(ctx, database.GetOffersByQueryParams{
		CONCAT: query,
		Limit:  offerPageSize,
		Offset: (page - 1) * offerPageSize,
	})
	if err != nil {
		return nil, err
	}
	for _, dbOffer := range dbOffers {
		relativeTime := RelativeTime(dbOffer.CreatedAt)
		offers = append(offers, shared.Offer{
			ID:              dbOffer.ID,
			Title:           dbOffer.Title,
			About:           dbOffer.About,
			Requirements:    dbOffer.Requirements,
			Benefits:        dbOffer.Benefits,
			Status:          dbOffer.Status,
			CompanyName:     dbOffer.CompanyName,
			CompanyID:       dbOffer.CompanyID,
			CompanyImageURL: dbOffer.CompanyImageUrl,
			MinWage:         dbOffer.MinWage,
			MaxWage:         dbOffer.MaxWage,
			RelativeTime:    relativeTime,
		})
	}
	return offers, nil
}

func (mysql *MysqlStorage) GetOffersByUser(ctx context.Context, userID string, page int32) ([]shared.Offer, error) {
	offers, err := mysql.Queries.GetOffersByUser(ctx, database.GetOffersByUserParams{
		ID:     userID,
		Limit:  offerPageSize,
		Offset: (page - 1) * offerPageSize,
	})
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
			ID:              offer.ID,
			Title:           offer.Title,
			About:           offer.About,
			Requirements:    offer.Requirements,
			Benefits:        offer.Benefits,
			Status:          offer.Status,
			CompanyName:     offer.CompanyName,
			CompanyID:       offer.CompanyID,
			CompanyImageURL: offer.CompanyImageUrl,
			MinWage:         offer.MinWage,
			MaxWage:         offer.MaxWage,
			RelativeTime:    relativeTime,
		})
	}
	return res, nil
}

func (mysql *MysqlStorage) SelectParticipatedOffers(ctx context.Context, userID string, page int32) ([]shared.Offer, error) {
	dbOffers, err := mysql.Queries.GetParticipatedOffers(ctx, database.GetParticipatedOffersParams{
		UserID: userID,
		Limit:  offerPageSize,
		Offset: (page - 1) * offerPageSize,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return []shared.Offer{}, nil
		}
		return []shared.Offer{}, err
	}
	offers := []shared.Offer{}
	for _, dbOffer := range dbOffers {
		offer := shared.Offer{
			ID:              dbOffer.ID,
			Title:           dbOffer.Title,
			About:           dbOffer.About,
			Requirements:    dbOffer.Requirements,
			Benefits:        dbOffer.Benefits,
			Status:          dbOffer.Status,
			CompanyName:     dbOffer.CompanyName,
			CompanyID:       dbOffer.CompanyID,
			CompanyImageURL: dbOffer.CompanyImageUrl,
			MinWage:         dbOffer.MinWage,
			MaxWage:         dbOffer.MaxWage,
			RelativeTime:    RelativeTime(dbOffer.CreatedAt),
		}
		offers = append(offers, offer)
	}
	return offers, nil
}

func (mysql *MysqlStorage) SelectOffers(ctx context.Context, params shared.OfferQueryParams) ([]shared.Offer, error) {
	if params.Query != "" {
		return mysql.GetOffersByQuery(ctx, params.Query, params.Page)
	}
	if params.CompanyID != "" {
		return mysql.GetOffersByCompany(ctx, params.CompanyID, params.Page)
	}
	if params.UserID != "" {
		return mysql.GetOffersByUser(ctx, params.UserID, params.Page)
	}
	return mysql.GetOffers(ctx, params.Page)
}

func RelativeTime(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	if duration < 0 {
		duration = -duration
		if duration < time.Minute {
			return "en unos segundos"
		} else if duration < time.Hour {
			return fmt.Sprintf("en %d minutos", int(duration.Minutes()))
		} else if duration < 24*time.Hour {
			return fmt.Sprintf("en %d horas", int(duration.Hours()))
		} else {
			days := int(duration.Hours() / 24)
			return fmt.Sprintf("en %d días", days)
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

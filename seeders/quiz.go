package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/kw3a/spotted-server/seeders/internal/database"
)

var description1 = `
	About DEUNA 🧡

We are a rapidly growing startup that simplifies global payments and powers next generation commerce in a single platform. With our products we've consolidated hundreds of payment solutions in a single integration, harness an intuitive payment orchestration method and centralize payment reconciliation.

We are currently present all across LATAM and looking for exceptional talent to join our team and continue revolutionizing the world of payments! 🚀

We are a dynamic tech team committed to creating, developing, and implementing microservices improvements tailored to meet the needs of our clients. As a Golang Developer, you will play a key role in shaping the future of our software solutions.

Responsibilities:

Create, test, and maintain applications and services using Golang.
Improve the performance of existing software by refactoring code as needed.
Design the architecture for new applications or services, ensuring scalability, maintainability, and security.
Connect different services and components through RESTful APIs, gRPC, or other protocols.
Write and maintain unit and integration tests to ensure software quality.
Keep code, architecture, and system feature documentation up-to-date.
Participate in code reviews to maintain a high level of quality and share knowledge with the team.
Work collaboratively with other developers, designers, and business team members to achieve project goals.
Implement security best practices to safeguard information and system resources.

Qualifications:

Bachelor's degree in Computer Science, Engineering, or a related field.
Proven 2 - 3+ years of experience as a Backend Developer with expertise in Golang.
Strong understanding of software architecture principles.
Experience with RESTful APIs, gRPC, and other integration protocols.
Proficient in writing automated tests.
Excellent collaboration and communication skills.
Familiarity with code review processes.
Knowledge of security best practices in software development.
	`

var description2 = `
Launchpad, a people-first technology company, is a leader in North America´s rapidly growing tech sector. Through two solutions, Launchpad supports its clients with digital transformation:

PaasportTM, our iPaaS solution, streamlines software integration and automates workflows. 
Nearshore Staff Augmentation, our managed IT staffing service, connects top IT talent across various geographical regions, bringing industry expertise to leading clients. 

Based in Vancouver, Canada, our operational footprint spans across North and South America, with a second headquarters in Santiago, Chile.

In 2023, our unwavering dedication to innovation garnered recognition as a Deloitte Technology Fast 50™ Program Company. Our clientele boasts industry leaders such as Walmart, GM, TIME Magazine, Salesforce, Tableau, Splunk, Bolt.com, Freedom House, and more.

At Launchpad, we genuinely care about our people as individuals. If you are looking for a team that values growth, drive, and passion for your craft, if you’re seeking a place to achieve your goals and dreams with fairness and integrity, then we’d love to hear from you.

Overview: we are looking for an Intermediate Data Engineer to support the SQL Consolidation program. This role will focus on executing the consolidation plan, providing hands-on technical expertise, and contributing to the overall success of the project.

Please note that 

This job will require your availability from Monday to Friday, 7.30 to 15.30 PST
Length of engagement estimated in 18 months
Estimated starting date: November 1st, 2024

Responsibilities:

Work closely with the Senior Data Engineer to implement the SQL consolidation plan. 
Perform hands-on technical tasks, including data migration, transformation, and optimization. 
Collaborate with the team to troubleshoot issues and ensure project milestones are met. 
Assist in documentation and reporting related to the consolidation efforts. 
Support knowledge transfer and training efforts as needed. 

Qualifications:

Solid experience with SQL Server, Azure, and other Microsoft data technologies. 
Strong technical skills in data migration, ETL, and database optimization. 
Ability to work collaboratively in a team environment. 
Strong analytical and problem-solving abilities. 
Effective communication skills, both written and verbal. 

Why work for Launchpad?

100% remote
People first culture
Excellent compensation in US Dollars
Hardware setup for working from home
Work with global teams and prominent brands based in North America, Europe, and Asia
Training allowances
Personal time off (PTO) for vacations, study leave, personal time, etc. 
...and more!

At Launchpad, we genuinely care about our people as individuals. If you are looking for a team that values growth, drive, and passion for your craft, if you’re seeking a place to achieve your goals and dreams with fairness and integrity, then you are the future of Launchpad. Launchpad is committed to fostering a diverse and representative workforce and an inclusive work environment where all employees are respected and treated equally.

Are you ready to elevate your career at Launchpad? We want to hear your story! Contact us today.
`

func (cfg *SeedersConfig) seedQuizes(authors []string) ([]string, error) {
	IDs := []string{
		uuid.New().String(),
		uuid.New().String(),
	}
	quizes := []database.SeedQuizParams{
		{
			ID:          IDs[0],
			Title:       "Junior Golang developer",
			Description: description1,
			Duration:    120,
			UserID:      authors[0],
		},
		{
			ID:          IDs[1],
			Title:       "Intermediate Data Engineer",
			Description: description2,
			Duration:    60,
			UserID:      authors[1],
		},
	}
	for _, quiz := range quizes {
		err := cfg.DB.SeedQuiz(cfg.Ctx, quiz)
		if err != nil {
			return nil, err
		}
	}
	fmt.Println("Quizzes seeded successfully")
	if err := SaveIDsToFile(IDs, "to_delete/quiz.txt"); err != nil {
		return nil, err
	}
	return IDs, nil
}

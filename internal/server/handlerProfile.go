package server

import (
	"net/http"

	"github.com/kw3a/spotted-server/internal/server/profiles"
)

func (DI *App) ProfilePageHandler() http.HandlerFunc {
	return profiles.CreateProfilePageHandler(
		DI.AuthService,
		DI.Templ,
		DI.Storage,
		profiles.GetProfilePageInput,
	)
}

func (DI *App) EducationRegisterHandler() http.HandlerFunc {
	return profiles.CreateRegisterEducationHandler(
		DI.Templ,
		DI.AuthService,
		DI.Storage,
		profiles.GetEducationRegisterInput,
	)
}

func (DI *App) EducationDeleteHandler() http.HandlerFunc {
	return profiles.CreateDeleteEducationHandler(
		DI.AuthService,
		DI.Storage,
		profiles.GetEducationDeleteInput,
	)
}

func (DI *App) ExperienceRegisterHandler() http.HandlerFunc {
	return profiles.CreateRegisterExperienceHandler(DI.Templ, DI.AuthService, DI.Storage, profiles.GetExperienceRegisterInput)
}

func (DI *App) ExperienceDeleteHandler() http.HandlerFunc {
	return profiles.CreateDeleteExperienceHandler(DI.AuthService, DI.Storage, profiles.GetExperienceDeleteInput)
}

func (DI *App) LinkRegisterHandler() http.HandlerFunc {
	return profiles.CreateRegisterLinkHandler(
		DI.Templ,
		DI.AuthService,
		DI.Storage,
		profiles.GetLinkRegisterInput,
	)
}

func (DI *App) LinkDeleteHandler() http.HandlerFunc {
	return profiles.CreateDeleteLinkHandler(
		DI.AuthService,
		DI.Storage,
		profiles.GetLinkDeleteInput,
	)
}

func (DI *App) LoginHandler() http.HandlerFunc {
	return profiles.CreateLoginHandler(
		DI.AuthType,
		DI.Storage,
		profiles.GetLoginInput,
		DI.Templ,
	)
}

func (DI *App) LoginPageHandler() http.HandlerFunc {
	return profiles.CreateLoginPageHandler(DI.AuthService, DI.Templ)
}

func (DI *App) LogoutHandler() http.HandlerFunc {
	return profiles.CreateLogoutHandler(DI.Storage, "/")
}

func (DI *App) ProfilePicHandler() http.HandlerFunc {
	return profiles.CreatePictureHandler(
		DI.Storage,
		DI.AuthService,
		&DI.Cld.Upload,
		profiles.GetProfilePicInput,
	)
}

func (DI *App) UserHandler() http.HandlerFunc {
	return profiles.CreateUserHandler(
		DI.AuthType,
		DI.Templ,
		DI.Storage,
		profiles.GetUserInput,
		"/profile/",
	)
}

func (DI *App) UserPageHandler() http.HandlerFunc {
	return profiles.CreateRegPageHandler(DI.AuthService, DI.Templ)
}

func (DI *App) SkillRegisterHandler() http.HandlerFunc {
	return profiles.CreateRegisterSkillHandler(DI.Templ, DI.AuthService, DI.Storage, profiles.GetSkillRegisterInput)
}

func (DI *App) SkillDeleteHandler() http.HandlerFunc {
	return profiles.CreateDeleteSkillHandler(DI.AuthService, DI.Storage, profiles.GetSkillDeleteInput)
}

func (DI *App) DescrUpdateHandler() http.HandlerFunc {
	return profiles.CreateDescUpdateHandler(
		DI.Templ,
		DI.AuthService,
		DI.Storage,
		profiles.GetDescUpdateInput,
	)
}

func (DI *App) UpdateEmail() http.HandlerFunc {
	return profiles.CreateUpdateEmailHandler(
		DI.Templ,
		DI.AuthService,
		DI.Storage,
		profiles.GetEmailUpdateInput,
	)
}

func (DI *App) UpdateCell() http.HandlerFunc {
	return profiles.CreateUpdateCellHandler(
		DI.Templ,
		DI.AuthService,
		DI.Storage,
		profiles.GetUpdateCellInput,
	)
}

// Package handlers contains HTTP handler functions for the web application.
package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/mcgigglepop/brilliant-inferno-ruby/server/internal/config"
	"github.com/mcgigglepop/brilliant-inferno-ruby/server/internal/forms"
	"github.com/mcgigglepop/brilliant-inferno-ruby/server/internal/models"
	"github.com/mcgigglepop/brilliant-inferno-ruby/server/internal/render"
)

// Repo is the repository used by the handlers.
var Repo *Repository

// Repository holds the application config and dependencies for handlers.
type Repository struct {
	App *config.AppConfig
}

// NewRepo creates a new Repository with the given app config.
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

// NewHandlers sets the global Repo variable to the provided repository.
func NewHandlers(r *Repository) {
	Repo = r
}

// ////////////////////////////////////////////////////////////
// ////////////////////////////////////////////////////////////
// /////////////////// GET REQUESTS ///////////////////////////
// ////////////////////////////////////////////////////////////
// ////////////////////////////////////////////////////////////

// LoginGet handles GET requests for the login page.
func (m *Repository) LoginGet(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{})
}

// RegisterGet handles GET requests for the registration page.
func (m *Repository) RegisterGet(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "register.page.tmpl", &models.TemplateData{})
}

// EmailVerificationGet handles GET requests for the email verification page.
// Redirects to login if no email is found in session.
func (m *Repository) EmailVerificationGet(w http.ResponseWriter, r *http.Request) {
	email := m.App.Session.GetString(r.Context(), "user_email")

	if email == "" {
		log.Println("No email found in session, cannot verify email")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	render.Template(w, r, "email-verification.page.tmpl", &models.TemplateData{})
}

// DashboardGet handles GET requests for the dashboard page.
func (m *Repository) DashboardGet(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "dashboard.page.tmpl", &models.TemplateData{})
}

//////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////
///////////////////// POST REQUESTS //////////////////////////
//////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////

// RegisterPost handles POST requests for user registration.
// Validates form, registers user with Cognito, and redirects appropriately.
func (m *Repository) RegisterPost(w http.ResponseWriter, r *http.Request) {
	if err := m.App.Session.RenewToken(r.Context()); err != nil {
		m.App.ErrorLog.Println("Session token renewal failed:", err)
	}

	err := r.ParseForm()
	if err != nil {
		m.App.ErrorLog.Println("ParseForm error:", err)
	}

	form := forms.New(r.PostForm)

	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		render.Template(w, r, "register.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	email := strings.TrimSpace(r.Form.Get("email"))
	password := r.Form.Get("password")

	userErr := m.App.CognitoClient.RegisterUser(r.Context(), email, password)
	if userErr != nil {
		m.App.ErrorLog.Println("Cognito RegisterUser failed:", userErr)
		m.App.Session.Put(r.Context(), "error", "Registration failed. Please try again.")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	// Store user email in session for verification step
	m.App.Session.Put(r.Context(), "user_email", email)

	m.App.Session.Put(r.Context(), "flash", "Registered successfully.")
	http.Redirect(w, r, "/email-verification", http.StatusSeeOther)
}

// EmailVerificationPost handles POST requests for email verification.
// Validates OTP form, confirms user with Cognito, and redirects appropriately.
func (m *Repository) EmailVerificationPost(w http.ResponseWriter, r *http.Request) {
	if err := m.App.Session.RenewToken(r.Context()); err != nil {
		m.App.ErrorLog.Println("Session token renewal failed:", err)
	}

	email := m.App.Session.GetString(r.Context(), "user_email")

	if email == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		m.App.ErrorLog.Println("ParseForm error:", err)
	}

	form := forms.New(r.PostForm)
	form.Required("otpFirst", "otpSecond", "otpThird", "otpFourth", "otpFifth", "otpSixth")

	if !form.Valid() {
		log.Printf("[DEBUG] Form validation failed: %+v", form.Errors)
		render.Template(w, r, "email-verification.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	// Concatenate OTP digits from form fields
	otpCode := strings.TrimSpace(
		r.Form.Get("otpFirst") +
			r.Form.Get("otpSecond") +
			r.Form.Get("otpThird") +
			r.Form.Get("otpFourth") +
			r.Form.Get("otpFifth") +
			r.Form.Get("otpSixth"),
	)

	_, err := m.App.CognitoClient.ConfirmUser(r.Context(), email, otpCode)
	if err != nil {
		m.App.ErrorLog.Printf("Cognito ConfirmUser failed: %v", err)
		m.App.Session.Put(r.Context(), "error", "Email verification failed. Please try again.")
		http.Redirect(w, r, "/email-verification", http.StatusSeeOther)
		return
	}

	// Remove user email from session after successful verification
	m.App.Session.Remove(r.Context(), "user_email")
	m.App.Session.Put(r.Context(), "flash", "Email Verified.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// LoginPost handles POST requests for user login.
// Validates form, logs in user with Cognito, and sets session tokens.
func (m *Repository) LoginPost(w http.ResponseWriter, r *http.Request) {
	if err := m.App.Session.RenewToken(r.Context()); err != nil {
		m.App.ErrorLog.Println("Session token renewal failed:", err)
	}

	err := r.ParseForm()
	if err != nil {
		m.App.ErrorLog.Println("ParseForm error:", err)
	}

	form := forms.New(r.PostForm)

	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	email := strings.TrimSpace(r.Form.Get("email"))
	password := r.Form.Get("password")

	auth_response, userErr := m.App.CognitoClient.Login(r.Context(), email, password)
	if userErr != nil {
		m.App.ErrorLog.Println("Cognito Login failed:", userErr)
		m.App.Session.Put(r.Context(), "error", "Login failed. Please try again.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sub, err := m.App.CognitoClient.ExtractSubFromToken(r.Context(), auth_response.IdToken)

	if err != nil {
		// handle error
	}

	// Store user ID and tokens in session
	m.App.Session.Put(r.Context(), "user_id", sub)
	m.App.Session.Put(r.Context(), "id_token", auth_response.IdToken)
	m.App.Session.Put(r.Context(), "access_token", auth_response.AccessToken)
	m.App.Session.Put(r.Context(), "refresh_token", auth_response.RefreshToken)

	m.App.Session.Put(r.Context(), "flash", "login successfully.")
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

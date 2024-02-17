package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"snippetbox.volcanoeyes.net/internal/models"
	"snippetbox.volcanoeyes.net/internal/validator"
)

type snippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type userPasswordUpdateForm struct {
	CurrentPassword     string `form:"current_password"`
	NewPassword         string `form:"new_password"`
	ConfirmPassword     string `form:"confirm_password"`
	validator.Validator `form:"-"`
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {

	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Snippets = snippets
	app.render(w, http.StatusOK, "home.html", data)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {

	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Snippet = snippet

	app.render(w, http.StatusOK, "view.html", data)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = snippetCreateForm{
		Expires: 365,
	}
	app.render(w, http.StatusOK, "create.html", data)
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {

	var form snippetCreateForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "create.html", data)
		return
	}

	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(w, http.StatusOK, "signup.html", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "Please enter a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "Your password must at least 8 characters")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "This email is already in use")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
			return
		}
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, http.StatusOK, "login.html", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "Please enter your email")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "Please enter a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "Please enter your password")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Invalid username or password")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	// currently if we move out of the login page the lastUrl token data will persist. what we really want is to remove this token when the user clicks off the page. Idea: add a middleware to delete the redirect token for the relevant handlers.
	lastUrl := app.sessionManager.PopString(r.Context(), "redirectPathAfterLogin")
	if lastUrl != "" {
		http.Redirect(w, r, lastUrl, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
	}
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "Successfully logged out!")
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (app *application) userAccount(w http.ResponseWriter, r *http.Request) {
	id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	user, err := app.users.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		} else {
			app.serverError(w, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.User = user
	app.render(w, http.StatusOK, "account.html", data)
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, http.StatusOK, "about.html", data)
}

func (app *application) passwordUpdate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userPasswordUpdateForm{}
	app.render(w, http.StatusOK, "update_password.html", data)
}

func (app *application) passwordUpdatePost(w http.ResponseWriter, r *http.Request) {
	var form userPasswordUpdateForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.CurrentPassword), "current_password", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.NewPassword), "new_password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.NewPassword, 8), "new_password", "Your new password must at least 8 characters")
	form.CheckField(validator.NotBlank(form.ConfirmPassword), "confirm_password", "This field cannot be blank")
	form.CheckField(form.NewPassword == form.ConfirmPassword, "confirm_password", "Passwords do not match")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "update_password.html", data)
		return
	}

	id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	err = app.users.UpdatePassword(id, form.CurrentPassword, form.NewPassword)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("The current password that you have entered is incorrect")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "update_password.html", data)
			return
		}
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Password change successful!")
	http.Redirect(w, r, "/account/view", http.StatusSeeOther)
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

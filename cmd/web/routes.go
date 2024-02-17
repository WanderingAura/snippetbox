package main

import (
	"net/http"

	"snippetbox.volcanoeyes.net/ui"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	// defines the router use our own notFound error, msg when there's no matching route
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})
	// makes a handler which serves HTTP requests with the contents of the file system
	fileServer := http.FileServer(http.FS(ui.Files))
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	router.HandlerFunc(http.MethodGet, "/ping", ping)

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	// routes that remove the redirect token
	dynamicRemoveToken := dynamic.Append(app.removeRedirectTokenData)

	router.Handler(http.MethodGet, "/", dynamicRemoveToken.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/about", dynamicRemoveToken.ThenFunc(app.about))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamicRemoveToken.ThenFunc(app.snippetView))
	router.Handler(http.MethodGet, "/user/signup", dynamicRemoveToken.ThenFunc(app.userSignup))
	router.Handler(http.MethodPost, "/user/signup", dynamicRemoveToken.ThenFunc(app.userSignupPost))

	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))

	protected := dynamic.Append(app.requireAuthentication)

	router.Handler(http.MethodGet, "/snippet/create", protected.ThenFunc(app.snippetCreate))
	router.Handler(http.MethodPost, "/snippet/create", protected.ThenFunc(app.snippetCreatePost))
	router.Handler(http.MethodGet, "/account/view", protected.ThenFunc(app.userAccount))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))
	router.Handler(http.MethodGet, "/account/password/update", protected.ThenFunc(app.passwordUpdate))
	router.Handler(http.MethodPost, "/account/password/update", protected.ThenFunc(app.passwordUpdatePost))

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	return standard.Then(router)
}

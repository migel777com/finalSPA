package main

import (
	"expvar"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/allBots", app.requirePermission("bots:admin", app.listBotsHandler))

	router.HandlerFunc(http.MethodPost, "/v1/bots", app.requirePermission("bots:write", app.createBotHandler))
	router.HandlerFunc(http.MethodGet, "/v1/bots/:id", app.requirePermission("bots:read", app.showBotHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/bots/:id", app.requirePermission("bots:write", app.updateBotHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/bots/:id", app.requirePermission("bots:write", app.deleteBotHandler))


	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodGet, "/checkToken/:token", app.checkTokenHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
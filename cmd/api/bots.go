package main

import (
	"errors"
	"finalSPA/internal/data"
	"finalSPA/internal/validator"
	"fmt"
	"net/http"
	"strconv"
)

func (app *application) createBotHandler(w http.ResponseWriter, r *http.Request) {
	auth_key := app.contextGetUser(r).ID

	var input struct {
		Owner_id int64 `json:"owner_id"`
		Bot_name string `json:"bot_name"`
		Bot_token string `json:"bot_token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	bot := &data.Bot{
		Owner_id: auth_key,
		Bot_name: input.Bot_name,
		Bot_token: input.Bot_token,
	}

	status, err := app.models.Bots.CheckToken(input.Bot_token)
	if err!=nil{
		app.notFoundResponse(w,r)
	}

	if status{
		bot.Bot_token_valid = true
	}

	v := validator.New()
	if data.ValidateBot(v, bot); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Bots.Insert(bot)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/bots/%d", bot.ID))
	err = app.writeJSON(w, http.StatusCreated, envelope{"bot": bot}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) checkTokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := app.readIDParam(r)


	status, err := app.models.Bots.CheckToken(strconv.FormatInt(token, 10))

	if err!=nil{
		app.notFoundResponse(w,r)
	}

	if status{

	}
}

func (app *application) showBotHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	auth_key := app.contextGetUser(r).ID
	owner_id, err := app.models.Bots.Get_Owner_Id(id)
	if err!=nil{
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if owner_id!=auth_key{
		app.notPermittedResponse(w, r)
	}


	bot, err := app.models.Bots.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"bot": bot}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}


func (app *application) updateBotHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	auth_key := app.contextGetUser(r).ID
	owner_id, err := app.models.Bots.Get_Owner_Id(id)
	if err!=nil{
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if owner_id!=auth_key{
		app.notPermittedResponse(w, r)
	}

	bot, err := app.models.Bots.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}


	var input struct {
		Bot_name *string `json:"bot_name"`
		Bot_token *string `json:"bot_token"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Bot_name != nil {
		bot.Bot_name = *input.Bot_name
	}
	if input.Bot_token != nil {
		bot.Bot_token = *input.Bot_token
	}


	v := validator.New()
	if data.ValidateBot(v, bot); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Bots.Update(bot)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"bot": bot}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteBotHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	auth_key := app.contextGetUser(r).ID
	owner_id, err := app.models.Bots.Get_Owner_Id(id)
	if err!=nil{
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if owner_id!=auth_key{
		app.notPermittedResponse(w, r)
	}


	err = app.models.Bots.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "bot successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listBotsHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Title string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)

	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")

	input.Filters.SortSafelist = []string{"id", "owner_id", "bot_name", "bot_token"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	bots, metadata, err := app.models.Bots.GetAll(input.Title, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Include the metadata in the response envelope.
	err = app.writeJSON(w, http.StatusOK, envelope{"bots": bots, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}



package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"finalSPA/internal/validator"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Bot struct {
	ID              int64     `json:"id"`
	Owner_id        int64     `json:"owner_id"`
	CreatedAt       time.Time `json:"-"`
	Bot_name        string    `json:"bot_name"`
	Bot_token       string    `json:"bot_token"`
	Bot_token_valid bool      `json:"Bot_token_valid"`
	Bot_isactive    bool      `json:"isActive"`
}

type status struct {
	Ok bool `json:"ok"`
}

type result struct {
	ID                          int64  `json:"id"`
	Is_bot                      bool   `json:"is_bot"`
	First_name                  string `json:"first_name"`
	Username                    string `json:"username"`
	Can_join_groups             bool   `json:"can_join_groups"`
	Can_read_all_group_messages bool   `json:"can_read_all_group_messages"`
	Supports_inline_queries     bool   `json:"supports_inline_queries"`
}

type checkTokenRequest struct {
	status status
	result result
}

type BotModel struct {
	DB *sql.DB
}

func ValidateBot(v *validator.Validator, bot *Bot) {
	v.Check(bot.Bot_name != "", "bot_name", "must be provided")
	v.Check(len(bot.Bot_name) <= 32, "bot_name", "must not be more than 32 bytes long")
	v.Check(len(bot.Bot_name) >= 6, "bot_name", "must be more than 6 bytes")
}

func (b BotModel) CheckToken(token string) (bool, error){
	var result map[string]interface{}
	res, err := http.Get("https://api.telegram.org/bot"+token+"/getMe")
	if err != nil {
		return false, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return false, err
	}
	for _, value := range result {
		if value==true{
			return true, nil
		}else{
			return false, nil
		}
	}

	return false, nil
}

func (b BotModel) Insert(bot *Bot) error {
	query := `INSERT INTO bots (owner_id, bot_name, bot_token)
				VALUES ($1, $2, $3)
				RETURNING id, created_at`

	args := []interface{}{bot.Owner_id, bot.Bot_name, bot.Bot_token}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return b.DB.QueryRowContext(ctx, query, args...).Scan(&bot.ID, &bot.CreatedAt)
}

func (b *BotModel) Get(id int64) (*Bot, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `SELECT id, owner_id, created_at, bot_name, bot_token, bot_token_valid, bot_isactive
				FROM bots
				WHERE id = $1`

	var bot Bot

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := b.DB.QueryRowContext(ctx, query, id).Scan(
		&bot.ID,
		&bot.Owner_id,
		&bot.CreatedAt,
		&bot.Bot_name,
		&bot.Bot_token,
		&bot.Bot_token_valid,
		&bot.Bot_isactive,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &bot, nil
}

func (b *BotModel) Get_Owner_Id(id int64) (int64, error) {
	if id < 1 {
		return 0, ErrRecordNotFound
	}
	query := `SELECT owner_id
				FROM bots
				WHERE id = $1`

	var owner int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := b.DB.QueryRowContext(ctx, query, id).Scan(
		&owner,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, ErrRecordNotFound
		default:
			return 0, err
		}
	}
	return owner, nil
}


func (b BotModel) GetAll(bot_name string, filters Filters) ([]*Bot, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, bot_name, bot_token, bot_token_valid, bot_isactive
		FROM bots
		WHERE (to_tsvector('simple', bot_name) @@ plainto_tsquery('simple', $1) OR $1 = '')
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{bot_name, filters.limit(), filters.offset()}

	rows, err := b.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	bots := []*Bot{}

	for rows.Next() {

		var bot Bot

		err := rows.Scan(
			&bot.ID,
			&bot.Owner_id,
			&bot.CreatedAt,
			&bot.Bot_name,
			&bot.Bot_token,
			&bot.Bot_token_valid,
			&bot.Bot_isactive,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		bots = append(bots, &bot)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return bots, metadata, nil
}


func (b BotModel) Update(bot *Bot) error {
	query := `UPDATE bots
				SET bot_name = $1, bot_token = $2
				WHERE id = $3
				RETURNING id`

	args := []interface{}{
		bot.Bot_name,
		bot.Bot_token,
		bot.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := b.DB.QueryRowContext(ctx, query, args...).Scan(&bot.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (b *BotModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM bots
				WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := b.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

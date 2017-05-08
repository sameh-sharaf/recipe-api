package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type DBManager struct {
	db *sql.DB
}

func InitConnection(host, user, pass, dbname, port string) (*DBManager, error) {
	connString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, dbname)
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return &DBManager{}, err
	}

	return &DBManager{db}, nil
}

func (dbManager *DBManager) ExecQuery(query string) (*sql.Rows, error) {
	res, err := dbManager.db.Query(query)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (dbManager *DBManager) ExecInsertRecipeQuery(name string, prep_time int, difficulty int8, vegeterian bool, createdAt time.Time) error {
	query := `
		INSERT INTO app.recipes (name, prep_time, difficulty, vegeterian, createdat, updatedat)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := dbManager.db.Query(query, name, prep_time, difficulty, vegeterian, createdAt.Format(time.RFC3339), createdAt.Format(time.RFC3339))
	return err
}

func (dbManager *DBManager) ExecInsertRateQuery(recipeID int64, rate int8, createdAt time.Time) error {
	query := `
		INSERT INTO app.rates (recipeID, rate, createdat)
		VALUES ($1, $2, $3)
	`
	_, err := dbManager.db.Query(query, recipeID, rate, createdAt.Format(time.RFC3339))
	return err
}

func (dbManager *DBManager) ExecUpdateQuery() error {
	return nil
}

func (dbManager *DBManager) ExecDeleteQuery(table, idKey, idVal string) error {
	query := fmt.Sprintf(`
		DELETE FROM %s
		WHERE %s = $1
	`, table, idKey)
	_, err := dbManager.db.Query(query, idVal)

	return err
}

func (dbManager *DBManager) EscapeQuotes(s string) string {
	return strings.Replace(s, "'", "''", -1)
}

func (dbManager *DBManager) InsertUser(username, fullName, passwordHash string) error {
	createdAt := time.Now().UTC().Format(time.RFC3339)
	query := `
		INSERT INTO app.users (username, fullName, passwordHash, createdAt)
		VALUES ($1, $2, $3, $4)
	`
	_, err := dbManager.db.Query(query, username, fullName, passwordHash, createdAt)
	return err
}

func (dbManager *DBManager) InsertUserSession(session Session) error {
	query := `
		INSERT INTO app.usersessions (sessionKey, userID, LoginTime)
		VALUES ($1, $2, $3)
	`
	_, err := dbManager.db.Query(query, session.SessionKey, session.User.ID, session.LoginTime.Format(time.RFC3339))
	return err
}

func (dbManager *DBManager) DeleteUserSessionByID(sessionKey string) error {
	query := "DELETE FROM app.usersessions WHERE sessionKey = $1;"
	_, err := dbManager.db.Query(query, sessionKey)
	return err
}

func (dbManager *DBManager) DeleteExpiredUserSessions(t time.Time) error {
	query := "DELETE FROM app.usersessions WHERE LoginTime < $1;"
	_, err := dbManager.db.Query(query, t.Format(time.RFC3339))
	return err
}

func (dbManagager *DBManager) GetUserActiveSessions(sessionKey string, maxLifeTime int64) (Session, error) {
	session := Session{}
	query := `SELECT a.id, a.username, a.fullname, b.sessionkey, b.LoginTime
						FROM app.users a
						INNER JOIN app.usersessions b
						ON a.id = b.userid
						WHERE a.isdisabled = FALSE
							AND b.sessionkey = $1
							AND b.LoginTime + $2 * interval '1 second' > CURRENT_TIMESTAMP;
	`
	res, err := dbManagager.db.Query(query, sessionKey, maxLifeTime)
	if err != nil {
		return session, err
	}

	if res.Next() {
		res.Scan(&session.User.ID, &session.User.Username, &session.User.Fullname, &session.SessionKey, &session.LoginTime)
	} else {
		return session, fmt.Errorf("could not find active session for token: %s", sessionKey)
	}

	return session, nil
}

func (dbManager *DBManager) GetUser(username string) (User, error) {
	user := User{}
	query := `SELECT id, username, fullname, passwordHash
						FROM app.users
						WHERE username = $1
							AND isdisabled = FALSE;
	`
	res, err := dbManager.db.Query(query, username)
	if err != nil {
		return user, err
	}

	if res.Next() {
		res.Scan(&user.ID, &user.Username, &user.Fullname, &user.PasswordHash)
	} else {
		return user, fmt.Errorf("user not found: %s", username)
	}

	return user, nil
}

func (dbManager *DBManager) GetRecipes(recipeID int64, items, page int) ([]Recipe, error) {
	whereClause := ""
	limitClause := ""
	offsetClause := ""

	if recipeID > 0 {
		whereClause = " WHERE a.id = $1"
	} else {
		if items > 0 {
			limitClause = fmt.Sprintf("LIMIT %v", items)
			if page > 0 {
				page--
			}
			offsetClause = fmt.Sprintf("OFFSET %v", items*page)
		}
	}

	query := fmt.Sprintf(`
			SELECT a.id, a.name, a.prep_time, a.difficulty, a.vegeterian, a.createdat, a.updatedat, COALESCE(AVG(b.rate), 0) AS rating
			FROM app.recipes a
			LEFT OUTER JOIN app.rates b
			ON a.id = b.recipeID
			%s
			GROUP BY 1, 2, 3, 4, 5
			ORDER BY a.createdat DESC
			%s %s;
		`, whereClause, limitClause, offsetClause)

	var err error
	var rows *sql.Rows
	if recipeID > 0 {
		rows, err = dbManager.db.Query(query, recipeID)
	} else {
		rows, err = dbManager.db.Query(query)
	}
	if err != nil {
		return nil, err
	}

	recipes := []Recipe{}
	for rows.Next() {
		recipe := Recipe{}
		err = rows.Scan(&recipe.ID, &recipe.Name, &recipe.PrepTime, &recipe.Difficulty, &recipe.Vegeterian, &recipe.CreatedAt, &recipe.UpdatedAt, &recipe.Rating)
		if err != nil {
			return nil, err
		}

		recipes = append(recipes, recipe)
	}

	return recipes, nil
}

func (dbManager *DBManager) GetRecipesByFilters(filters string) ([]Recipe, error) {
	recipes := []Recipe{}
	whereClause := fmt.Sprintf("WHERE %s", filters)
	query := fmt.Sprintf(`
		SELECT * FROM
			(SELECT a.id, a.name, a.prep_time, a.difficulty, a.vegeterian, a.createdat, a.updatedat, COALESCE(AVG(b.rate), 0) AS rating
			FROM app.recipes a
			LEFT OUTER JOIN app.rates b
			ON a.id = b.recipeID
			GROUP BY 1, 2, 3, 4, 5) a
			%s
			ORDER BY createdat DESC;
		`, whereClause)
	res, err := dbManager.db.Query(query)
	if err != nil {
		return recipes, err
	}

	for res.Next() {
		recipe := Recipe{}
		res.Scan(&recipe.ID, &recipe.Name, &recipe.PrepTime, &recipe.Difficulty, &recipe.Vegeterian, &recipe.CreatedAt, &recipe.UpdatedAt, &recipe.Rating)
		recipes = append(recipes, recipe)
	}

	return recipes, nil
}

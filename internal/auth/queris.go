package auth

const (
	queryInsertUser = `
		INSERT INTO users (id, login, password_hash)
		VALUES($1, $2, $3)
	`
	queryInsertBalance = `
		INSERT INTO balances (user_id, current, withdrawn)
		VALUES ($1, $2, $3)
	`
	queryGetUserByLogin = `
		SELECT id, password_hash
		FROM users WHERE login = $1
	`
)

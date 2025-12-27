package repository

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
	queryInsertOrder = `
		INSERT INTO orders (number, user_id)
		VALUES ($1, $2)
	`
	querySelectOrder = `
		SELECT user_id
		FROM orders
		WHERE number = $1
	`
	queryUpdateOrderStatus = `
		UPDATE orders
		SET status = $1, accrual $2
		WHERE number = $3 AND status NOT IN ('PROCESSED', 'INVALID')
		RETURNING user_id
	`
	queryUpdateBalance = `
		UPDATE balances
		SET current = current + $1
		WHERE user_id = $2
	`
	querySelectBalance = `
		SELECT current, withdrawn
		FROM balances
		WHERE user_id = $1
	`
	queryUpdateBalanceDecrease = `
		UPDATE balances
		SET current = current - $1, withdrawn = withdrawn + $1
		WHERE user_id = $2 AND current >= $1
	`
	queryInsertWithdrawal = `
		INSERT INTO withdrawal (user_id, order_number, sum)
		VALUES ($1, $2, $3)
	`
	querySelectWithdrawalsByUser = `
		SELECT order_number, sum, processed_at
		FROM withdrawal
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`
	querySelectOrdersBase = `
		SELECT number, status, uploaded_at
		FROM orders WHERE 
	`
)

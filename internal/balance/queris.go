package balance

const (
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
		INSERT INTO withdrawal (id, user_id, order_number, sum)
		VALUES ($1, $2, $3, $4)
	`
	querySelectWithdrawalsByUser = `
		SELECT order_number, sum, processed_at
		FROM withdrawal
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`
)

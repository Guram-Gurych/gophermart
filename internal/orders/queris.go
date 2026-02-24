package orders

const (
	queryInsertOrder = `
		INSERT INTO orders (number, user_id)
		VALUES ($1, $2)
		ON CONFLICT (number) DO UPDATE SET number = EXCLUDED.number
		RETURNING user_id, (xmax = 0) AS is_inserted;
	`
	queryUpdateOrderStatusandAccrual = `
		UPDATE orders
		SET status = $1, accrual = $2
		WHERE number = $3 AND status NOT IN ('PROCESSED', 'INVALID')
		RETURNING user_id;
	`
	queryUpdateOrderStatus = `
		UPDATE orders 
		SET status = $1 
		WHERE number = ANY($2) AND status IN ('NEW', 'REGISTERED', 'PROCESSING')
	`
	queryUpdateBalance = `
		UPDATE balances
		SET current = current + $1
		WHERE user_id = $2
	`
	querySelectOrdersBase = `
		SELECT number, status, accrual, uploaded_at
		FROM orders 
	`
)

package storage

const insertStmt = `WITH e AS (
			INSERT INTO short_url (original_url, short_url, user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (original_url) DO NOTHING
		RETURNING short_url
		)
		SELECT short_url, 100000
		FROM e
		UNION ALL
		SELECT short_url, 100001
		FROM short_url
		WHERE original_url=$1`

const getOriginalURLStmt = `select original_url, deleted from short_url where short_url=$1`
const getUserURL = `select original_url, short_url from short_url where user_id=$1`
const batchInsert = `INSERT INTO short_url(id, short_url, original_url, user_id) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO UPDATE SET id = excluded.id RETURNING id;`

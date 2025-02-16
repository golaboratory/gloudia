-- name: FindWishById :one
select *
from m_wishes mu 
where id = $1;

-- name: GetWishAll :many
select *
from m_wishes mu;
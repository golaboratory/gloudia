-- name: FindUserById :one
select *
from m_user mu 
where id = $1;


-- name: GetUserAll :many
select *
from m_user mu;

-- name: TryLogin :one
select *
-- name: FindUserById :one
select *
from m_user mu 
where id = $1;


-- name: GetUserAll :many
select *
from m_user mu;

-- name: TryLogin :one
select *
from m_user	mu
where login_id = $1
and mu.password_hash = $2
and mu.is_deleted = false;
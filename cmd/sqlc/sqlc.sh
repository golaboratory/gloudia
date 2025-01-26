if [ -d $1 ]; then
    docker run --rm -v $1:/src -w /src sqlc/sqlc generate
else
    echo "Specify a directory path containing sqlc.(yaml|yml) or sqlc.json as the first argument."
fi

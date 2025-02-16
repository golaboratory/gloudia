go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.3.0

oapi-codegen --config=./codegen.config.sample.yml ./openapi.yaml


if [ -d $1 ]; then
docker run --rm -v $1:/local openapitools/openapi-generator-cli generate \
    -i /local/openapi.yaml \
    -g go-gin-server  \
    -o /local/server
else
    echo "Specify a directory path containing openapi.(yaml|yml) or sqlc.json as the first argument."
fi

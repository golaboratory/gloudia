docker run -p 8081:8080 -d -v ".:/tmp" -e SWAGGER_FILE=/tmp/openapi.yml swaggerapi/swagger-editor

start http://localhost:8081/
# Phoenix price API

### Prerequisite

- Install `swagger-combine`:
```
npm install --save swagger-combine
```
- Install `statik`:
```
go install github.com/rakyll/statik@v0.1.7
```

### Deployment
Follow below instructions:
- Setup env:
```
export READONLY_URL=<postgresql_url>
```
- Building binary:
```
go build main.go
```
- Running grpc server and host public OpenAPI gateway:
```
./main
```
Then you can query your own API through [localhost:5050](localhost:5050) or use OpenAPI UI at [localhost:5050/public](localhost:5050/public)

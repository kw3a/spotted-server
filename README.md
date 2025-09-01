## Installation
Install judge0 locally following: https://github.com/judge0/judge0/blob/master/CHANGELOG.md#deployment-procedure and judge0/docker-compose-judge0.yml

Development tools (gosec, staticcheck, air):
```shell
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go install github.com/air-verse/air@latest
```

Database
```shell
docker compose . up -d
chmod +x ./scripts/migrateup.sh
./scripts/migrateup.sh 
```

Run gosec tests
```shell
gosec ./...
```

Run staticcheck linter
```shell
staticcheck ./...
```

Run unit tests
```shell
go test ./...
```

Run seeders
```shell
chmod +x ./scripts/seed.sh 
./scripts/seed.sh
```

Run the server
```shell
air
```


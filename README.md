## Installation
Install judge0 locally using https://github.com/judge0/judge0/blob/master/CHANGELOG.md#deployment-procedure and judge0/docker-compose-judge0.yml

Install Mysql locally using Docker
```shell
docker compose . up -d
```

Create database
```shell
chmod +x ./scripts/migrateup.sh
./scripts/migrateup.sh 
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


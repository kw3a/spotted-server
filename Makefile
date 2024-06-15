CONTAINER_NAME := spotted-mysql
PASSWORD_FILE := password.txt
DBPASSWORD := $(shell cat $(PASSWORD_FILE))

initDB:
	docker compose up -d

containerDB:
	docker exec -it $(CONTAINER_NAME) sh

enterDB:
	docker exec -it $(CONTAINER_NAME) mysql -u root -p$(DBPASSWORD) -D spotted-db

stopDB:
	docker stop $(CONTAINER_NAME)

deleteDB: 
	docker rm $(CONTAINER_NAME) 

createSeederDirectory:
	mkdir -p seeders/to_delete

createSeederFiles:
	touch seeders/to_delete/example.txt \
		seeders/to_delete/language_problem.txt \
		seeders/to_delete/language.txt \
		seeders/to_delete/problem.txt \
		seeders/to_delete/quiz.txt \
		seeders/to_delete/test_case.txt \
		seeders/to_delete/user.txt

createPasswordFile:
	@test -e password.txt || echo "root" > password.txt

seed: createSeederDirectory createSeederFiles
	./scripts/seed.sh

upMigration:
	./scripts/migrateup.sh

downMigration:
	./scripts/migratedown.sh

run:
	./scripts/run.sh

showtime: initDB createPasswordFile sleep10 upMigration seed run

sleep10:
	sleep 10
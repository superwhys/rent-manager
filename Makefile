.PHONY: generate build_sql

generate:
	go generate ./internal/ent
	go generate ./api

build_sql:
	atlas schema inspect --url "ent://internal/ent/schema?dialect=mysql" --dev-url "docker://mysql/8/dev" --format '{{ sql . "  " }}' > sql/init.sql
.PHONY: generate 

generate:
	go generate ./internal/infra/persistence/ent
	go generate ./api

build_sql:
	atlas schema inspect --url "ent://internal/infra/persistence/ent/schema?dialect=mysql" --dev-url "docker://mysql/8/dev" --format '{{ sql . "  " }}' > sql/init.sql
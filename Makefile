# Halora Land — safe docker compose helpers.
#
# NEVER run `docker compose down -v` / `docker volume rm` — it deletes the
# pgdata named volume and wipes all database data (AHSP would need a
# re-import). `make down` below is the safe full stop: volumes are kept.

.PHONY: up stop start down logs ps build

## Build images and start the stack
up:
	docker compose up -d --build

## Rebuild changed images and restart (fast iteration)
build:
	docker compose up -d --build

## Pause containers (data + containers kept, resumes instantly with make start)
stop:
	docker compose stop

## Resume paused containers
start:
	docker compose start

## Full stop + remove containers/networks — volumes are KEPT, data safe
down:
	docker compose down

## Follow logs (all services)
logs:
	docker compose logs -f --tail=100

## Container status
ps:
	docker compose ps

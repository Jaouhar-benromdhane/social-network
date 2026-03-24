#!/usr/bin/env bash

set -u

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

CONTAINERS=("social-network-backend" "social-network-frontend")

compose_cmd=()

detect_compose_cmd() {
  if docker compose version >/dev/null 2>&1; then
    compose_cmd=(docker compose -f docker-compose.yml)
    return 0
  fi

  if command -v docker-compose >/dev/null 2>&1; then
    compose_cmd=(docker-compose -f docker-compose.yml)
    return 0
  fi

  echo "Erreur: docker compose/docker-compose introuvable."
  return 1
}

run_compose() {
  "${compose_cmd[@]}" "$@"
}

print_status() {
  echo
  echo "Etat des conteneurs social-network:"
  docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "social-network-(backend|frontend)" || echo "Aucun conteneur trouve."
  echo
}

start_server() {
  detect_compose_cmd || return 1

  echo "Demarrage backend + frontend..."
  if run_compose up -d backend frontend >/dev/null 2>&1; then
    echo "Services demarres via docker compose."
  else
    echo "docker compose indisponible/erreur, tentative fallback avec docker start..."
    local started_any=0
    local name=""
    for name in "${CONTAINERS[@]}"; do
      if docker start "$name" >/dev/null 2>&1; then
        echo "- $name demarre"
        started_any=1
      else
        echo "- impossible de demarrer $name"
      fi
    done

    if [[ "$started_any" -eq 0 ]]; then
      echo "Echec: aucun conteneur n'a pu etre demarre."
      return 1
    fi
  fi

  print_status
  echo "Frontend: http://localhost:3000"
  echo "Backend:  http://localhost:8080/api/health"
}

stop_server() {
  detect_compose_cmd || return 1

  echo "Arret backend + frontend..."
  if run_compose stop backend frontend >/dev/null 2>&1; then
    echo "Services arretes via docker compose."
  else
    echo "docker compose indisponible/erreur, tentative fallback in-place..."
    local name=""
    for name in "${CONTAINERS[@]}"; do
      if docker ps --format "{{.Names}}" | grep -Fxq "$name"; then
        if docker exec "$name" sh -lc "kill -TERM 1" >/dev/null 2>&1; then
          sleep 1
          if docker ps --format "{{.Names}}" | grep -Fxq "$name"; then
            docker exec "$name" sh -lc "kill -KILL 1" >/dev/null 2>&1 || true
            sleep 1
          fi

          if docker ps --format "{{.Names}}" | grep -Fxq "$name"; then
            echo "- impossible d'arreter $name"
          else
            echo "- $name arrete"
          fi
        else
          echo "- impossible d'arreter $name"
        fi
      else
        echo "- $name est deja arrete"
      fi
    done
  fi

  print_status
}

restart_server() {
  stop_server || true
  sleep 1
  start_server
}

usage() {
  echo "Usage: ./server.sh {start|stop|restart|status}"
}

main() {
  if [[ $# -lt 1 ]]; then
    usage
    return 1
  fi

  case "$1" in
    start)
      start_server
      ;;
    stop)
      stop_server
      ;;
    restart)
      restart_server
      ;;
    status)
      print_status
      ;;
    *)
      usage
      return 1
      ;;
  esac
}

main "$@"
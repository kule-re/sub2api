#!/usr/bin/env bash
# Update an existing Compose installation; never bootstrap a new database.
# Usage: bash hajimi-deploy.sh /opt/sub2api ghcr.io/kule-re/hajimi@sha256:... [registry-user]
# With registry-user, read the short-lived registry token from stdin.
set -Eeuo pipefail
umask 077

fail() { printf 'hajimi: %s\n' "$*" >&2; exit 1; }
[[ $# -ge 2 && $# -le 3 ]] || fail 'expected deployment directory, immutable image, optional registry user'
deploy_dir=$1
image=$2
registry_user=${3:-}
[[ $deploy_dir == /* && -d $deploy_dir ]] || fail 'deployment directory must already exist and be absolute'
[[ $image =~ ^ghcr\.io/kule-re/hajimi@sha256:[a-f0-9]{64}$ ]] || fail 'expected a hajimi GHCR image pinned by SHA256 digest'
for tool in docker flock realpath mktemp; do command -v "$tool" >/dev/null || fail "missing command: $tool"; done
docker compose version >/dev/null
deploy_dir=$(realpath -- "$deploy_dir")
cd "$deploy_dir"
[[ -f .env ]] || fail 'existing .env is missing; refusing to initialize a new installation'

exec 9> .hajimi-deploy.lock
flock -n 9 || fail 'another deployment is already running'

container=sub2api
working_dir=$(docker inspect "$container" --format '{{index .Config.Labels "com.docker.compose.project.working_dir"}}')
[[ -n $working_dir && -d $working_dir ]] || fail 'running container has no valid Compose working directory'
[[ $(realpath -- "$working_dir") == "$deploy_dir" ]] || fail 'running container belongs to a different deployment directory'
project=$(docker inspect "$container" --format '{{index .Config.Labels "com.docker.compose.project"}}')
[[ $project =~ ^[a-z0-9][a-z0-9_-]*$ ]] || fail 'running container has no valid Compose project name'
config_files=$(docker inspect "$container" --format '{{index .Config.Labels "com.docker.compose.project.config_files"}}')
[[ -n $config_files && $config_files != '<no value>' ]] || fail 'running container has no Compose file list'

override="$deploy_dir/compose.hajimi.yml"
compose=(docker compose --project-directory "$deploy_dir" --project-name "$project" --env-file "$deploy_dir/.env")
IFS=',' read -r -a originals <<< "$config_files"
base_files=()
for file in "${originals[@]}"; do
  [[ -f $file ]] || fail "original Compose file missing: $file"
  file=$(realpath -- "$file")
  # Do not accumulate duplicate generated overrides on subsequent deployments.
  if [[ $file != "$override" ]]; then
    compose+=(-f "$file")
    base_files+=("$file")
  fi
done
[[ ${#base_files[@]} -gt 0 ]] || fail 'original base Compose file is missing'

candidate=
registry_dir=
cleanup() {
  [[ -z $candidate ]] || rm -f -- "$candidate"
  [[ -z $registry_dir ]] || rm -rf -- "$registry_dir"
}
trap cleanup EXIT
trap 'printf "hajimi: deployment failed; no automatic image or database rollback. Check the server and retained backups.\n" >&2' ERR

if [[ -n $registry_user ]]; then
  registry_dir=$(mktemp -d /tmp/hajimi-registry.XXXXXXXX)
  export DOCKER_CONFIG="$registry_dir"
  IFS= read -r registry_token || fail 'missing registry token on stdin'
  [[ -n $registry_token ]] || fail 'empty registry token'
  printf '%s' "$registry_token" | docker login ghcr.io --username "$registry_user" --password-stdin >/dev/null
  unset registry_token
fi

# Download before backups/recreation so registry failures leave the app untouched.
docker pull "$image"
candidate=$(mktemp "$deploy_dir/.hajimi-compose.XXXXXXXX.yml")
printf 'services:\n  sub2api:\n    image: %s\n    pull_policy: never\n' "$image" > "$candidate"
"${compose[@]}" -f "$candidate" config --quiet

mkdir -p backups
stamp=$(date -u +%Y%m%dT%H%M%SZ)
backup_dir=$(mktemp -d "$deploy_dir/backups/hajimi-${stamp}.XXXXXXXX")
cp -- .env "$backup_dir/.env"
for i in "${!base_files[@]}"; do
  cp -- "${base_files[$i]}" "$backup_dir/compose-${i}-$(basename "${base_files[$i]}")"
done
if [[ -f $override ]]; then cp -- "$override" "$backup_dir/compose.hajimi.previous.yml"; fi
old_image=$(docker inspect "$container" --format '{{.Image}}')
old_tag="hajimi:rollback-${stamp}-${backup_dir##*.}"
docker image tag "$old_image" "$old_tag"
printf '%s\n' "$old_image" > "$backup_dir/old-image-id.txt"
printf '%s\n' "$old_tag" > "$backup_dir/old-image-tag.txt"
printf '%s\n' "$image" > "$backup_dir/new-image.txt"

# Use the running database's own client, database and user. Never print secrets.
"${compose[@]}" exec -T postgres sh -c \
  'PGPASSWORD="$POSTGRES_PASSWORD" pg_dump -h 127.0.0.1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc' \
  > "$backup_dir/database.dump"
[[ -s $backup_dir/database.dump ]] || fail 'database backup is empty'
"${compose[@]}" exec -T postgres pg_restore --list < "$backup_dir/database.dump" > /dev/null
# Works for both bind-mounted and named-volume application data.
docker exec "$container" tar -C /app/data -czf - . > "$backup_dir/app-data.tar.gz"
[[ -s $backup_dir/app-data.tar.gz ]] || fail 'application data backup is empty'
printf 'hajimi: backup complete: %s\n' "$backup_dir"

# Preserve the original project, files, environment, networks and volume paths.
mv -- "$candidate" "$override"
candidate=
"${compose[@]}" -f "$override" up -d --no-deps --no-build --pull never \
  --wait --wait-timeout 180 sub2api
expected_id=$(docker image inspect "$image" --format '{{.Id}}')
actual_id=$(docker inspect "$container" --format '{{.Image}}')
[[ $actual_id == "$expected_id" ]] || fail 'running container does not match the requested image'
"${compose[@]}" -f "$override" ps sub2api
printf 'hajimi: deployment healthy: %s\n' "$image"

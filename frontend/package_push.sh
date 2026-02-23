#!/usr/bin/env bash
set -euo pipefail

ACTION="${1:-push}"

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$project_dir"

DEPLOY_SSH_TARGET="${DEPLOY_SSH_TARGET:-}"
DEPLOY_HOST="${DEPLOY_HOST:-114.132.245.76}"
DEPLOY_USER="${DEPLOY_USER:-root}"
DEPLOY_PORT="${DEPLOY_PORT:-}"
DEPLOY_PATH="${DEPLOY_PATH:-/111workspace/news}"
DEPLOY_RESTART_CMD="${DEPLOY_RESTART_CMD:-}"
DEPLOY_SSH_ARGS="${DEPLOY_SSH_ARGS:-}"

resolve_ssh_target() {
  local ssh_target="${DEPLOY_SSH_TARGET}"
  if [ -z "${ssh_target}" ]; then
    if [ -z "${DEPLOY_HOST}" ] || [ -z "${DEPLOY_USER}" ]; then
      echo "Deploy failed: DEPLOY_SSH_TARGET or DEPLOY_HOST/DEPLOY_USER is required."
      exit 1
    fi
    ssh_target="${DEPLOY_USER}@${DEPLOY_HOST}"
    DEPLOY_PORT="${DEPLOY_PORT:-22}"
  fi
  echo "${ssh_target}"
}

declare -a SSH_ARGS=()
RSYNC_SSH_CMD="ssh"

set_transport_args() {
  local extra_ssh_args="${1:-}"

  SSH_ARGS=()
  RSYNC_SSH_CMD="ssh"

  if [ -n "${DEPLOY_SSH_ARGS}" ]; then
    SSH_ARGS+=( ${DEPLOY_SSH_ARGS} )
    RSYNC_SSH_CMD="ssh ${DEPLOY_SSH_ARGS}"
  fi

  if [ -n "${extra_ssh_args}" ]; then
    SSH_ARGS+=( ${extra_ssh_args} )
    RSYNC_SSH_CMD="${RSYNC_SSH_CMD} ${extra_ssh_args}"
  fi

  if [ -n "${DEPLOY_PORT}" ]; then
    SSH_ARGS+=(-p "${DEPLOY_PORT}")
    RSYNC_SSH_CMD="${RSYNC_SSH_CMD} -p ${DEPLOY_PORT}"
  fi
}

should_retry_without_proxy() {
  local output="${1:-}"
  if [ -n "${DEPLOY_SSH_ARGS}" ]; then
    return 1
  fi
  case "${output}" in
    *"127.0.0.1 port 7897"*|*"kex_exchange_identification"*|*"Connection reset by peer"*)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

ssh_exec() {
  local ssh_target="$1"
  local remote_cmd="$2"

  local output=""
  local status=0

  set_transport_args ""
  set +e
  set +u
  output="$(ssh ${SSH_ARGS[@]+"${SSH_ARGS[@]}"} "${ssh_target}" "${remote_cmd}" 2>&1)"
  status=$?
  set -u
  set -e

  if [ "${status}" -ne 0 ] && should_retry_without_proxy "${output}"; then
    set_transport_args "-o ProxyCommand=none"
    set +e
    set +u
    output="$(ssh ${SSH_ARGS[@]+"${SSH_ARGS[@]}"} "${ssh_target}" "${remote_cmd}" 2>&1)"
    status=$?
    set -u
    set -e
  fi

  if [ -n "${output}" ]; then
    echo "${output}"
  fi
  return "${status}"
}

rsync_upload() {
  local src="$1"
  local dst="$2"

  local output=""
  local status=0

  set_transport_args ""
  set +e
  output="$(rsync -az -e "${RSYNC_SSH_CMD}" "${src}" "${dst}" 2>&1)"
  status=$?
  set -e

  if [ "${status}" -ne 0 ] && should_retry_without_proxy "${output}"; then
    set_transport_args "-o ProxyCommand=none"
    set +e
    output="$(rsync -az -e "${RSYNC_SSH_CMD}" "${src}" "${dst}" 2>&1)"
    status=$?
    set -e
  fi

  if [ -n "${output}" ]; then
    echo "${output}"
  fi
  return "${status}"
}

build_artifact() {
  if [ ! -f "./package.sh" ]; then
    echo "package.sh not found"
    exit 1
  fi
  local artifact
  artifact="$(bash ./package.sh | tail -n 1)"
  if [ ! -f "${artifact}" ]; then
    echo "Artifact not found: ${artifact}"
    exit 1
  fi
  echo "${artifact}"
}

latest_artifact() {
  local latest=""
  if [ -d "./release" ]; then
    latest="$(ls -t ./release/frontend-dist-*.tar.gz 2>/dev/null | head -n 1 || true)"
  fi
  if [ -z "${latest}" ]; then
    echo ""
  else
    echo "${latest}"
  fi
}

deploy_artifact() {
  local artifact="${ARTIFACT:-}"
  if [ -z "${artifact}" ]; then
    artifact="$(latest_artifact)"
  fi
  if [ -z "${artifact}" ]; then
    echo "No artifact found. Run: ./package_push.sh build"
    exit 1
  fi
  if [ ! -f "${artifact}" ]; then
    echo "Artifact not found: ${artifact}"
    exit 1
  fi

  local ssh_target
  ssh_target="$(resolve_ssh_target)"

  if [ -z "${DEPLOY_PATH}" ]; then
    echo "Deploy failed: DEPLOY_PATH is required."
    exit 1
  fi

  local remote_release_dir="${DEPLOY_PATH}/release"
  local remote_artifact="${remote_release_dir}/$(basename "${artifact}")"

  echo "Preparing remote directory ${ssh_target}:${DEPLOY_PATH}/ ..."
  ssh_exec "${ssh_target}" "mkdir -p \"${remote_release_dir}\""

  echo "Uploading artifact to ${ssh_target}:${remote_artifact} ..."
  if command -v rsync >/dev/null 2>&1; then
    rsync_upload "${artifact}" "${ssh_target}:${remote_artifact}"
  else
    if [ -n "${DEPLOY_PORT}" ]; then
      scp ${DEPLOY_SSH_ARGS} -P "${DEPLOY_PORT}" "${artifact}" "${ssh_target}:${remote_artifact}"
    else
      scp ${DEPLOY_SSH_ARGS} "${artifact}" "${ssh_target}:${remote_artifact}"
    fi
  fi

  echo "Extracting dist/ to ${ssh_target}:${DEPLOY_PATH}/dist ..."
  ssh_exec "${ssh_target}" "bash -lc 'set -e; mkdir -p \"${DEPLOY_PATH}\"; rm -rf \"${DEPLOY_PATH}/dist\"; tar -xzf \"${remote_artifact}\" -C \"${DEPLOY_PATH}\"; test -d \"${DEPLOY_PATH}/dist\"'"

  if [ -n "${DEPLOY_RESTART_CMD}" ]; then
    echo "Running restart command on remote..."
    ssh_exec "${ssh_target}" "bash -lc '${DEPLOY_RESTART_CMD}'"
  fi

  echo "Deploy successful."
}

print_help() {
  echo "Usage:"
  echo "  ./package_push.sh (default: push - build & deploy)"
  echo "  ./package_push.sh build"
  echo "  ./package_push.sh deploy"
  echo "  ./package_push.sh push"
  echo ""
  echo "Deploy env:"
  echo "  DEPLOY_SSH_TARGET (optional, ssh config host or user@host)"
  echo "  DEPLOY_HOST (default: 114.132.245.76)"
  echo "  DEPLOY_USER (default: root)"
  echo "  DEPLOY_PORT (default: 22 when DEPLOY_SSH_TARGET is empty)"
  echo "  DEPLOY_PATH (default: /111workspace/news/frontend)"
  echo "  DEPLOY_RESTART_CMD (optional, e.g. sudo systemctl reload nginx)"
  echo "  DEPLOY_SSH_ARGS (optional, extra ssh options)"
  echo "  ARTIFACT (optional, local tar.gz path for deploy)"
}

case "${ACTION}" in
  build)
    build_artifact
    ;;
  deploy)
    deploy_artifact
    ;;
  push)
    ARTIFACT="$(build_artifact)"
    export ARTIFACT
    deploy_artifact
    ;;
  help|-h|--help)
    print_help
    ;;
  *)
    print_help
    exit 1
    ;;
esac

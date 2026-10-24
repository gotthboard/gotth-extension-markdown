# shellcheck shell=sh
rg_reject_matches() {
  rg_program=$1
  violation_message=$2
  shift 2
  if "$rg_program" "$@"; then
    rg_status=0
  else
    rg_status=$?
  fi
  case $rg_status in
    0) echo "$violation_message" >&2; return 1 ;;
    1) return 0 ;;
    *) echo "policy scanner failed with status $rg_status" >&2; return "$rg_status" ;;
  esac
}

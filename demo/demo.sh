type_and_run_all() {
  local cmds=("$@")
  for cmd in "${cmds[@]}"; do
    clear
    printf '>> '
    printf '%s' "$cmd" | pv -qL 7
    echo
    sleep 1
    eval "$cmd"
    sleep 3
  done
}

commands=(
  "diffee left right"
  "diffee left right --crc32 --info"
  "diffee left right --size --info --no-empty"
  "diffee left right --plain --files --left-orphans --single-quotes"
  "diffee left right --plain --files --right-orphans --swap --double-quotes"
)

type_and_run_all "${commands[@]}"


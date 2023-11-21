#compdef tlc3

# This script inspired by https://github.com/urfave/cli
# NOTE: Complex completions such as flag combination checks are not supported

_tlc3() {
  local -a opts
  local cur
  cur="${words[-1]}"

  if [[ "$cur" == "-"* ]]; then
    opts=($(_CLI_ZSH_AUTOCOMPLETE_HACK=1 ${words[@]:0:${#words[@]}-1} ${cur} --generate-bash-completion))
  else
    opts=($(_CLI_ZSH_AUTOCOMPLETE_HACK=1 ${words[@]:0:${#words[@]}-1} --generate-bash-completion))
  fi

  if [[ "${opts[1]}" != "" ]]; then
    _describe 'values' opts
  else
    _files
  fi
}

command -v tlc3 >/dev/null 2>&1 && compdef _tlc3 tlc3

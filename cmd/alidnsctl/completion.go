package main

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

var CompltionCommand = &cli.Command{
	Name:      "completion",
	Usage:     "generate shell completion script",
	UsageText: CompletionUseage,
	Subcommands: []*cli.Command{
		{Name: "bash", Action: GenerateBashCompletion},
		{Name: "zsh", Action: GenerateZshCompletion},
	},
}

var CompletionUseage = `
  On Bash
  	source <(alidnsctl completion bash)
  Or 
  	alidnsctl completion bash | sudo tee /etc/bash_completion.d/alidnsctl

  
  On Zsh
  	source <(alidnsctl completion zsh)
  Or
  	alidnsctl completion zsh > "${fpath[1]}/_alidnsctl"
  
  
  On Powershell
  	Add-Content $PROFILE "if (Get-Command alidnsctl -ErrorAction SilentlyContinue) {
  	alidnsctl completion powershell | Out-String | Invoke-Expression
  	}"
  OR
  	alidnsctl completion powershell >> $PROFILE
`

var BashCompletionScript = `#! /bin/bash

: ${PROG:=alidnsctl}

_cli_bash_autocomplete() {
  if [[ "${COMP_WORDS[0]}" != "source" ]]; then
    local cur opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    if [[ "$cur" == "-"* ]]; then
      opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} ${cur} --generate-bash-completion )
    else
      opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-bash-completion )
    fi
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    return 0
  fi
}

complete -o bashdefault -o default -o nospace -F _cli_bash_autocomplete $PROG
unset PROG

`

var ZshCompletionScript = `#compdef $PROG

_cli_zsh_autocomplete() {
  local -a opts
  local cur
  cur=${words[-1]}
  if [[ "$cur" == "-"* ]]; then
    opts=("${(@f)$(${words[@]:0:#words[@]-1} ${cur} --generate-bash-completion)}")
  else
    opts=("${(@f)$(${words[@]:0:#words[@]-1} --generate-bash-completion)}")
  fi

  if [[ "${opts[1]}" != "" ]]; then
    _describe 'values' opts
  else
    _files
  fi
}

compdef _cli_zsh_autocomplete $PROG
`

func GenerateBashCompletion(*cli.Context) error {
	fmt.Print(strings.ReplaceAll(BashCompletionScript, "$PROG", "alidnsctl"))
	return nil
}

func GenerateZshCompletion(*cli.Context) error {
	fmt.Print(strings.ReplaceAll(ZshCompletionScript, "$PROG", "alidnsctl"))
	return nil
}

#
# bash completion file for docker-machine commands
#
# This script provides completion of:
#  - commands and their options
#  - machine names
#  - filepaths
#
# To enable the completions either:
#  - place this file in /etc/bash_completion.d
#  or
#  - copy this file to e.g. ~/.docker-machine-completion.sh and add the line
#    below to your .bashrc after bash completion features are loaded
#    . ~/.docker-machine-completion.sh
#

_docker-machine-active() {
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--swarm --help" -- "${cur}"))
    else
        COMPREPLY=()
    fi
}

_docker-machine-config() {
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--swarm --help" -- "${cur}"))
    else
        COMPREPLY=($(compgen -W "$(docker-machine ls -q)" -- "${cur}"))
    fi
}

_docker-machine-create() {
    # cheating, b/c there are approximately one zillion options to create
    COMPREPLY=($(compgen -W "$(docker-machine create --help | grep '^   -' | sed 's/^   //; s/[^a-z0-9-].*$//')" -- "${cur}"))
}

_docker-machine-env() {
    case "${prev}" in
        --shell)
            # What are the options for --shell?
            COMPREPLY=()
            ;;
        *)
            if [[ "${cur}" == -* ]]; then
                COMPREPLY=($(compgen -W "--swarm --shell --unset --help" -- "${cur}"))
            else
                COMPREPLY=($(compgen -W "$(docker-machine ls -q)" -- "${cur}"))
            fi
    esac
}

# See docker-machine-wrapper.bash for the use command
_docker-machine-use() {
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--swarm --unset --help" -- "${cur}"))
    else
        COMPREPLY=($(compgen -W "$(docker-machine ls -q)" -- "${cur}"))
    fi
}

_docker-machine-inspect() {
    case "${prev}" in
        -f|--format)
            COMPREPLY=()
            ;;
        *)
            if [[ "${cur}" == -* ]]; then
                COMPREPLY=($(compgen -W "--format --help" -- "${cur}"))
            else
                COMPREPLY=($(compgen -W "$(docker-machine ls -q)" -- "${cur}"))
            fi
            ;;
    esac
}

_docker-machine-ip() {
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--help" -- "${cur}"))
    else
        COMPREPLY=($(compgen -W "$(docker-machine ls -q)" -- "${cur}"))
    fi
}

_docker-machine-kill() {
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--help" -- "${cur}"))
    else
        COMPREPLY=($(compgen -W "$(docker-machine ls -q)" -- "${cur}"))
    fi
}

_docker-machine-ls() {
    case "${prev}" in
        --filter)
            COMPREPLY=()
            ;;
        *)
            COMPREPLY=($(compgen -W "--quiet --filter --help" -- "${cur}"))
            ;;
    esac
}

_docker-machine-regenerate-certs() {
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--help --force" -- "${cur}"))
    else
        COMPREPLY=($(compgen -W "$(docker-machine ls -q)" -- "${cur}"))
    fi
}

_docker-machine-restart() {
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--help" -- "${cur}"))
    else
        COMPREPLY=($(compgen -W "$(docker-machine ls -q)" -- "${cur}"))
    fi
}

_docker-machine-rm() {
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--help --force" -- "${cur}"))
    else
        # For rm, it's best to be explicit
        COMPREPLY=()
    fi
}

_docker-machine-ssh() {
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--help" -- "${cur}"))
    else
        COMPREPLY=($(compgen -W "$(docker-machine ls -q)" -- "${cur}"))
    fi
}

_docker-machine-scp() {
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--help --recursive" -- "${cur}"))
    else
        _filedir
        # It would be really nice to ssh to the machine and ls to complete
        # remote files.
        COMPREPLY=($(compgen -W "$(docker-machine ls -q | sed 's/$/:/')" -- "${cur}") "${COMPREPLY[@]}")
    fi
}

_docker-machine-start() {
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--help" -- "${cur}"))
    else
        COMPREPLY=($(compgen -W "$(docker-machine ls -q)" -- "${cur}"))
    fi
}

_docker-machine-status() {
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--help" -- "${cur}"))
    else
        COMPREPLY=($(compgen -W "$(docker-machine ls -q)" -- "${cur}"))
    fi
}

_docker-machine-stop() {
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--help" -- "${cur}"))
    else
        COMPREPLY=($(compgen -W "$(docker-machine ls -q)" -- "${cur}"))
    fi
}

_docker-machine-upgrade() {
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--help" -- "${cur}"))
    else
        COMPREPLY=($(compgen -W "$(docker-machine ls -q)" -- "${cur}"))
    fi
}

_docker-machine-url() {
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--help" -- "${cur}"))
    else
        COMPREPLY=($(compgen -W "$(docker-machine ls -q)" -- "${cur}"))
    fi
}

_docker-machine-help() {
    if [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "--help" -- "${cur}"))
    else
        COMPREPLY=($(compgen -W "${commands[*]}" -- "${cur}"))
    fi
}

_docker-machine-docker-machine() {
    if [[ " ${wants_file[*]} " =~ " ${prev} " ]]; then
        _filedir
    elif [[ " ${wants_dir[*]} " =~ " ${prev} " ]]; then
        _filedir -d
    elif [[ "${cur}" == -* ]]; then
        COMPREPLY=($(compgen -W "${flags[*]} ${wants_dir[*]} ${wants_file[*]}" -- "${cur}"))
    else
        COMPREPLY=($(compgen -W "${commands[*]}" -- "${cur}"))
    fi
}

_docker-machine() {
    COMPREPLY=()
    local commands=(active config create env inspect ip kill ls regenerate-certs restart rm ssh scp start status stop upgrade url help)

    local flags=(--debug --native-ssh --help --version)
    local wants_dir=(--storage-path)
    local wants_file=(--tls-ca-cert --tls-ca-key --tls-client-cert --tls-client-key)

    # Add the use subcommand, if we have an alias loaded
    if [[ ${DOCKER_MACHINE_WRAPPED} = true ]]; then
        commands=("${commands[@]}" use)
    fi

    local cur prev words cword
    _get_comp_words_by_ref -n : cur prev words cword
    local i
    local command=docker-machine

    for (( i=1; i < ${cword}; ++i)); do
        local word=${words[i]}
        if [[ " ${wants_file[*]} ${wants_dir[*]} " =~ " ${word} " ]]; then
            # skip the next option
            (( ++i ))
        elif [[ " ${commands[*]} " =~ " ${word} " ]]; then
            command=${word}
        fi
    done

    local completion_func=_docker-machine-"${command}"
    if declare -F "${completion_func}" > /dev/null; then
        ${completion_func}
    fi

    return 0
}

complete -F _docker-machine docker-machine

#!/bin/bash
set -e

function help {
    cat <<EOF
Usage: $1 gitlab_ssh_port where_to_put_gitlab_ssh_key_file [other options for docker create]

Normally you will need root permission to run this script.
EOF
    exit 1
}

if [[ $1 == "" || $2 == "" ]]
then
    help $0
fi

port="$1"
shift
file="$1"
shift

touch "$file"
chown 998:998 "$file"
chmod 644 "$file"

exec docker create "$@" -p $port:22 -v "$file:/var/opt/gitlab/.ssh/authorized_keys" gitlab/gitlab-ce

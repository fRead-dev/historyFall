#!/bin/sh

VERSION="1.0"

##################################################################################3

NAME=$(git branch | grep '*')
HASH=$( (printf "commit %s\0" $(git cat-file commit HEAD | wc -c); git cat-file commit HEAD) | sha1sum )
HASH=$(echo $HASH | cut -c -8)

echo "$NAME [$VERSION.$HASH]" ': ' $(cat "$1") > "$1"

exit 0


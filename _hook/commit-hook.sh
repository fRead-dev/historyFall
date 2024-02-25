#!/bin/sh

VERSION="1.0"
fileConst="$PWD/pkg/module/historyConst.go"
dateNow=$(date +"%m-%d-%Y")


##################################################################################

NAME=$(git branch | grep '*')
HASH=$( (printf "commit %s\0" $(git cat-file commit HEAD | wc -c); git cat-file commit HEAD) | sha1sum )
HASH=$(echo $HASH | cut -c -8)

echo "$NAME [$VERSION.$HASH]" ': ' $(cat "$1") > "$1"

###################################

echo "package module" > "$fileConst"
echo "" >> "$fileConst"
echo "const constVersionHistoryFall string = \"$VERSION.$HASH\"" >> "$fileConst"
echo "const constDateUpdateHistoryFall = \"$dateNow\"" >> "$fileConst"

exit 0


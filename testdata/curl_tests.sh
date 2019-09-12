#!/bin/bash

RESULT=`curl --disable -s -X GET http://localhost:8080/v1/ping`
if [ "$RESULT" != "pong" ]; then
    echo "Actual  : '$RESULT'"
    echo "Expected: 'pong'"
    exit 1
fi
# $RESULT should be pong

urls=(
    "http://example.com" # Simple URL
    "https://example.com" # For Security
    "https://www.duckduckgo.com/" # Ends with a /
    "https://duckduckgo.com/?q=oh+the+humanity&t=ffsb&ia=web" # Includes parameters
    "https://www.google.com/search?hl=en&q=https%3A%2F%2Fduckduckgo.com%2F%3Fq%3Dlorem%2Bipsum%26t%3Dffsb%26ia%3Danswer" # Includes url encoding
    # TODO: Add the following tests
    #   Really super long
)

for url in "${urls[@]}"
do
    SHORT_URL=`curl --disable -s -X POST 'http://localhost:8080/v1/create' -d '{"longUrl":"'$url'"}'`
    LONG_URL=`curl --disable -s -w "%{redirect_url}\n" -o /dev/null -X GET $SHORT_URL`

    if [ "$LONG_URL" != "$url" ]; then
        echo "Actual  : '$LONG_URL'"
        echo "Expected: '$url'"
        exit 1
    fi
done

echo "Pass"

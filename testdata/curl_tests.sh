#!/bin/sh

curl -X GET http://localhost:8080/v1/ping

curl -X POST 'http://localhost:8080/v1/create' -d '{"longUrl":"https://www.urlencoder.org/"}'
curl -X POST 'http://localhost:8080/v1/create' -d '{"longUrl":"https://www.urlencoder.org/"}'
curl -X POST 'http://localhost:8080/v1/create' -d '{"longUrl":"https://example.com"}'
curl -X POST 'http://localhost:8080/v1/create' -d '{"longUrl":"https://www.duckduckgo.com/"}'
curl -X POST 'http://localhost:8080/v1/create' -d '{"longUrl":"https://duckduckgo.com/?q=oh+the+humanity&t=ffsb&ia=web"}'
curl -X POST 'http://localhost:8080/v1/create' -d '{"longUrl":"https://www.google.com/search?hl=en&q=https%3A%2F%2Fduckduckgo.com%2F%3Fq%3Dlorem%2Bipsum%26t%3Dffsb%26ia%3Danswer"}'

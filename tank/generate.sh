#!/bin/bash
> ammo.txt

for i in {1..10}; do
  body="{\"name\":\"book-$i\",\"author_ids\":[\"e6672056-49ee-4aba-a9f0-21813b2963a3\"]}"
  body_len=$(echo -n "$body" | wc -c)
  request=$(printf "POST /v1/library/book HTTP/1.1\r\nHost: 127.0.0.1\r\nContent-Type: application/json\r\nContent-Length: %d\r\n\r\n%s" "$body_len" "$body")
  total_len=$(echo -ne "$request" | wc -c)
  echo "$total_len POST_book" >> ammo.txt
  echo -ne "$request\n\n" >> ammo.txt
done
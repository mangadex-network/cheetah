## cheetah

A diverse MangaDex@Home client

## Quick Start

### Build

```bash
go build -ldflags="-s -w -extldflags '-static'" -o ./bin/cheetah ./cli/main.go
```

### Run (modes)

```bash
# stand-alone (with or without token validation)
./bin/cheetah --key=XXXXXXXX --port=44300 --cache=/var/mdath/cache

# proxy for a cluster (with or without ssl / token validation)
# e.g. load balancing, TLS termination, ...
./bin/cheetah proxy --key=XXXXXXXX --port=44300 --origins=http://192.168.0.38:8000,https://uploads.mangadex.org

# cache node in a cluster (without ssl / token validation)
# e.g. high performance image hosting
./bin/cheetah cache --port=8000 --upstream=https://uploads.mangadex.org --cache=/var/mdath/cache
```

## Development

Start local image server
```bash
# start local image server
go run ./cli/main.go --key=XXXXXXXX --port=44300 --cache=./test/cache --no-token-check
# test get image
curl --insecure 'https://127.0.0.1:44300/SbVLV10h4HZ56rE9a19BK3inEyiFBBipKqYMxKRgQwdYr_v8cSctYp6beEO495Zc86x1UJ48V95DtezIOGheZriAVm5WYx5LPiwOpXWAnuZed9HMZtCRaEK_D77rP_EmU5au6XcQbG54fJWW4kRbNpMidmNEOvbA8V8bpdGgGNXpwWAlSl_NaggYM7X1BxnC/data/8172a46adc798f4f4ace6663322a383e/B18-8ceda4f88ddf0b2474b1017b6a3c822ea60d61e454f7e99e34af2cf2c9037b84.png' > /dev/null
# benchmark
ab -n 2500 -c 50 'https://127.0.0.1:44300/SbVLV10h4HZ56rE9a19BK3inEyiFBBipKqYMxKRgQwdYr_v8cSctYp6beEO495Zc86x1UJ48V95DtezIOGheZriAVm5WYx5LPiwOpXWAnuZed9HMZtCRaEK_D77rP_EmU5au6XcQbG54fJWW4kRbNpMidmNEOvbA8V8bpdGgGNXpwWAlSl_NaggYM7X1BxnC/data/8172a46adc798f4f4ace6663322a383e/B18-8ceda4f88ddf0b2474b1017b6a3c822ea60d61e454f7e99e34af2cf2c9037b84.png'
```

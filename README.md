# Share

Share is a web service that permit to securely share files and secrets to anyone.

<br>

## Install from sources
```
git clone https://github.com/ggtrd/share.git
cd share
go mod tidy
```
```
go run share web
```

<br>

## Install with Docker

### Get docker-compose.yml
```
curl -O https://raw.githubusercontent.com/ggtrd/share/refs/heads/main/docker-compose.yml
```

```
docker compose up -d
```

<br>

## Use the CLI

> if runned with Docker:
> ```docker exec -it share-share-1 sh```

```
go run share help
```

<br>

## License
This project is licensed under the MIT License. See the LICENSE file for details.
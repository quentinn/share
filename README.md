# Share

Share is a web service that permit to securely share files and secrets to anyone


## Install from sources
```
git clone https://github.com/ggtrd/share.git
cd share
go mod tidy
```
```
go run share web
```

## Install with Docker

### Get docker-compose.yml
```
curl -O https://raw.githubusercontent.com/ggtrd/share/refs/heads/main/docker-compose.yml
```

```
docker compose up -d
```


## Use the CLI

if runned with Docker:
```
docker exec compose up -d
```

```
go run share help
```


## License
This project is licensed under the MIT License. See the LICENSE file for details.
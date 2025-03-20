# Share

Share is a web service that permit to securely share files and secrets to anyone.

<br>

## Install from sources
```
git clone git@github.com:ggtrd/share.git \
cd share \
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
> ```docker exec -it <container> sh```

```
go run share help
```

<br>

## Reverse proxy example

### Apache HTTP Server with authentication on share creation
```
sudo a2enmod ssl proxy proxy_http
```

/etc/apache/sites-available/001-share.conf
```
ServerName share.<domain>

<VirtualHost *:80>
        Redirect permanent / https://share.<domain>
</VirtualHost>

<VirtualHost *:443>
	SSLEngine on
	SSLCertificateFile      /etc/ssl/certs/<cert>.pem
	SSLCertificateKeyFile   /etc/ssl/private/<cert>.key

	# Allow everyone to get /share (= unlock share page)
	<Location "/share">
		AuthType none
		Satisfy any
	</Location>

	# Allow everyone to get /static (= style, images and JS files)
	<Location "/static">
		AuthType none
		Satisfy any
	</Location>

	# Require an authentication for everything else like /secret and /files (= create shares)
	# The authentication can be anything (basic auth, OIDC etc...)
	<Location "/">
		AuthType Basic
		AuthUserFile /etc/apache2/.htpasswd
		Require valid-user
	</Location>

	ProxyPreserveHost On
	ProxyRequests On
	ProxyPass / http://0.0.0.0:8080/
	ProxyPassReverse / http://0.0.0.0:8080/
</VirtualHost>
```

```
sudo a2ensite 001-share.conf
```

<br>

## Disclaimer
OpenPGP is used to cipher password when unlocking the password of a share (to avoid brute-force passwords attempts). It doesn't cipher anything else, please consider using HTTPS with a TLS certificate.

<br>

## License
This project is licensed under the MIT License. See the LICENSE file for details.


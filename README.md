# Go-XMPP-Upload

An implementation of a custom upload service for XEP-0363: HTTP File Upload written in Go. Its tested and documented for usage with the Ejabberd server and Postgres as database to store upload slots and file info.

Tested with Conversations (Android) and Gajim (with http-upload plugin).

## Environment Variables

Configuration is done through various environment variables. The Following vars are available:

| Variable                      | Description                                                                                                                                                                                               | Example                   | Required  |
|-----------------------------  |---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |------------------------   |---------- |
| EJABBERD_PORT_5222_TCP_ADDR   | Host- or IP-Address of the Ejabberd Server. Only this IP (and IPs listed in ALLOWED_IPS)  can request a upload slot. If you use Docker this Variable is filled  automatically through container linking.  | 172.17.0.2                |     ✓     |
| POSTGRES_PORT_5432_TCP_ADDR   | Host- or IP-Address of your Postgres server. If you use Docker this Variable is filled  automatically through container linking                                                                           | 172.17.0.3                |     ✓     |
| POSTGRES_USER                 | The Postgres user                                                                                                                                                                                         | postgres                  |     ✓     |
| POSTGRES_PASSWORD             | The Password for the given user                                                                                                                                                                           | hunter2                   |     ✓     |
| POSTGRES_DATABASE             | Postgres database name                                                                                                                                                                                    | xmpp-upload               |     ✓     |
| ALLOWED_IPS                   | Comma separated list of additionally allowed IP  addresses which can request upload slots.                                                                                                                | 127.0.0.1, 192.168.0.1    |     x     |
| PUT_GET_URL_HOST              | The host part of the PUT/GET URL which will  send to the Jabber Client                                                                                                                                    | http://yourdomain.tld     |     ✓     |
| XMPP_UPLOAD_LISTEN            | Listening string for the xmpp-upload http server. If you use Docker the port must be set to 8080  except if you change the Dockerfile to expose  another port                                             | :8080                     |     x     |
| UPLOADED_FILES_DIR            | Directory for storing uploaded files. Make sure this directory exists                                                                                                                                     | /opt/xmpp_uploads         |     ✓     |


## Preparations

### Checkout

Clone this Repository in you ```$GOPATH```


### Configuring with Ejabberd

Make sure to use a version of Ejabberd which supports mod_http_upload (e.g. >= 15.10).

Add the following to your ejabberd.yml config file in the modules part

```
  mod_http_upload: 
    service_url: "https://$PUBLIC_ACCESSABLE_XMPP_UPLOAD_HOST/slot"
```

Replace $PUBLIC_ACCESSABLE_XMPP_UPLOAD_HOST with your xmpp-upload listening host/ip and port and change ```https``` to ```http``` if you dont use ssl.


### Importing Postgres SQL scheme

Import file ```postgres-scheme.sql``` to your database using a shell with psql or with pgadmin


## Using with Docker


### Build Image

There is a ```Dockerfile``` in the repo root dir, so you can simply run the following from there:

```
docker build -t xmpp_upload:latest .
```


### Create Data Container

```
docker create --name xmpp-upload-data xmpp_upload:latest
```

### Run

```
docker run -d \
    --name "xmpp-upload-run" \
    --volumes-from xmpp-upload-data \
    --link ejabberd-run:postgresql \
    --link postgres-run:postgresql \
    -p 8080:8080 \
    -h 'yourhostname' \
    -e "POSTGRES_USER=postgres" \
    -e "POSTGRES_PASSWORD=hunter2" \
    -e "POSTGRES_DATABASE=xmpp-upload" \
    -e "PUT_GET_URL_HOST=http://yourdomain.tld" \
    -e "UPLOADED_FILES_DIR=/opt/xmpp-files"
    xmpp_upload:latest
```


To utilize SSL you can run xmpp-upload behind a nginx acting as SSL proxy otherwise the generated random url hashes could be sniffed.


## Using without Docker

Run the following from the root dir of the repo inside your $GOPATH

```
go install && \
env EJABBERD_PORT_5222_TCP_ADDR="$EJABBERD-HOST" \
POSTGRES_PORT_5432_TCP_ADDR="$POSTGRES-HOST" \
POSTGRES_USER="$POSTGRES_USER" \
POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
POSTGRES_DATABASE="xmpp_upload" \
ALLOWED_IPS="$ALLOWED-IPS" \
PUT_GET_URL_HOST="$PUT-GET-URL-HOST" \
XMPP_UPLOAD_LISTEN="$XMPP-UPLOAD-LISTEN" \
UPLOADED_FILES_DIR="/opt/xmpp-files" \
../../../../bin/go-xmpp-upload
```


## Backup

TODO

### Files

TODO

### Postgres data

TODO



Copyright (c) 2016 Tobias Hartwich (tobias.hartwich@gmail.com)


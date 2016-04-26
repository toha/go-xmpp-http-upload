# XMPP Upload


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


## Preparations

### Checkout

Clone this Repository in you $GOPATH

### Configuring with Ejabberd


### Importing Postgres SQL scheme

Import file ```postgres-scheme.sql``` to your database using a shell with psql or with pgadmin


## Build and run with Docker

### Build Image


### Create Data Container


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
    tha_xmpp_upload
```


## Using without Docker

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
../../../../bin/go-xmpp-upload
```

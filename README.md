# go-buildserver

## Build the server
```
docker build -t buildserver .
```

## Run the server after building
```
docker run -p 3000:3000 -v /home/pjotr/.ssh/:/data/.ssh/:ro -v /./example/config.yaml:/data/config.yaml -v /home/pjotr/.ssh/known_hosts:/root/.ssh/known_hosts:ro -e CONFIG_PATH=/app/example/config.yaml buildserver
```

## Run and build with docker compose
```
docker compose up
```
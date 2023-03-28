# go-buildserver
A very basic build server, written in Go. Why, because I needed a build server
which, didn't use a lot of ram, like the other solutions available.

What is not a lot of ram? Example setup uses less then 30MB of RAM inside the
given Docker container.

Documentation of the config file can be found in the docs folder: [docs/config_format.md](docs/config_format.md)

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
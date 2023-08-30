# Better-Manga-Proxy

Better Manga App is an open-source project aimed at simplifying the reading process. This repository is not an essential part of the application, but it can speed up the image loading process by caching the image.

## Quick Start

### Running with Docker

The easiest way to start with the server is by running it as a Docker container.

1. Create `docker-compose.yml`

The following file is an example of what the files should resemble or look like.

`docker-compose.yml`

```bash
version: "3.7"

services:
  better-manga-proxy:
    image: nohackjustnoobb/better-manga-proxy
    container_name: better-manga-proxy
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - ADDRESS=<BACKEND SERVER URL>
      # Remove the line below if you don't want to limit the cache size
      - MAX_CACHE_SIZE=1024 # unit - MB
```

2. Start the server

The following command will pull the docker image and start the server.

```bash
sudo docker-compose up -d
```

### Manual Setup

In order to run the server, it is essential to create a `.env` file. An example of the `.env` file is shown below:

`.env`

```python
ADDRESS=<BACKEND SERVER URL>

# optional
# remove it disabled limitation for cache size
MAX_CACHE_SIZE=1024 # unit - MB
```

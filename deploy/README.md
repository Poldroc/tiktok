# Deploy

You can **directly use** all the contents in this folder to deploy the **complete project**.

# Introduction

the shell file(`restart-service.sh`) is used for start/restart the specific service (**list: api user follow interaction video chat**)

```bash
sh restart-service.sh     # start all services
sh restart-service.sh api # start specific
```

- the container use **host** network
- you need to **move config.yaml** to the same dir as this file
- **Before starting the service**, you need to configure the corresponding environment (refer to the configuration in config.yaml).
- The script will **automatically** detect and delete the containers, no manual deletion is required.


# Quick use

1. move all the contents under this directory to your server.
2. complete the `config.yaml` (there are prompts inside the file).
3. if not launch env, use `docker-compose up -d` to launch.
4. enter this directory, simply execute `sh restart-service.sh`.

**You may need to set up an nginx reverse proxy to allow external requests.**
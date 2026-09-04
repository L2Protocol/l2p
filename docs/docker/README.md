## Docker Image

Included in this repo is a Dockerfile that you can launch an L2P node with for trying it out. Docker images are available on `ghcr.io/l2protocol/l2p`.

You can build the docker image with the following commands:
```bash
make docker
```

If your build machine has an ARM-based chip, like Apple silicon (M1), the image is built for `linux/arm64` by default. To build for `x86_64`, apply the --platform arg:

```bash
docker build --platform linux/amd64 -t l2p -f Dockerfile .
```

Before starting the docker container, get a copy of `config.toml` & `genesis.json` from the release: https://github.com/L2Protocol/l2p/releases, and make any necessary modification. Both files should be mounted into `/l2p/config` inside the container. Assuming they are under `./config` in your current working directory, you can start your docker container with the following command:
```bash
docker run -v $(pwd)/config:/l2p/config --rm --name l2p -it l2p
```

The container writes its chain data to `/data`, so mount a volume there if you want the data to survive a restart:
```bash
docker run -v $(pwd)/config:/l2p/config -v $(pwd)/data:/data --rm --name l2p -it l2p
```

You can also use `ETHEREUM OPTIONS` to overwrite settings in the configuration file
```bash
docker run -v $(pwd)/config:/l2p/config --rm --name l2p -it l2p --http.addr 0.0.0.0 --http.port 8545 --http.vhosts '*' --verbosity 3
```

If you need to open another shell, just do:
```bash
docker exec -it l2p /bin/bash
```

We also provide a `docker-compose` file for local testing.

To use the container in kubernetes, you can use a configmap or secret to mount the `config.toml` & `genesis.json` into the container
```bash
containers:
  - name: l2p
    image: l2p

    ports:
      - name: p2p
        containerPort: 31398
      - name: p2p-udp
        containerPort: 31398
        protocol: UDP
      - name: rpc
        containerPort: 8545
      - name: ws
        containerPort: 8546

    volumeMounts:
      - name: l2p-config
        mountPath: /l2p/config

  volumes:
    - name: l2p-config
      configMap:
        name: cm-l2p-config
```

Your configmap `l2p-config` should look like this:
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-l2p-config
data:
  config.toml: |
    ...

  genesis.json: |
    ...  

```

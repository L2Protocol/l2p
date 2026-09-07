useradd --system --no-create-home --shell /usr/sbin/nologin geth

mkdir /etc/geth
chown geth:geth -R /etc/geth

mkdir /var/lib/geth
chown geth:geth -R /var/lib/geth

wget -qO /usr/local/bin/geth https://github.com/L2Protocol/l2p/releases/latest/download/geth_linux
chmod +x /usr/local/bin/geth

wget -q https://static.l2protocol.com/config/geth/genesis.json

#Archive node
geth --datadir /var/lib/geth \
  --syncmode full \
  --gcmode archive \
  --networkid 12216 \
  --port 31398 \
  --discovery.port 31398 \
  --discovery.v5 \
  --config /etc/geth/config.toml \
  --nodekey /etc/geth/node.key \
  --txlookuplimit 0 \
  --history.logs 0 \
  --history.state 0 \
  --history.transactions 0 \
  init genesis.json

  #Validator node
  geth --datadir /var/lib/geth \
  --syncmode full \
  --networkid 12216 \
  --port 31398 \
  --discovery.port 31398 \
  --discovery.v5 \
  --config /etc/geth/config.toml \
  --nodekey /etc/geth/node.key \
  --txlookuplimit 0 \
  --history.logs 0 \
  --history.state 0 \
  --history.transactions 0 \
  init genesis.json
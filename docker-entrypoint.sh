#!/bin/bash
set -e

L2P_CONFIG=${L2P_HOME}/config/config.toml
L2P_GENESIS=${L2P_HOME}/config/genesis.json

# Init genesis state if the chain database does not exist yet
if [ ! -d "${DATA_DIR}/geth" ]; then
  geth --datadir "${DATA_DIR}" init "${L2P_GENESIS}"
fi

exec geth --config "${L2P_CONFIG}" --datadir "${DATA_DIR}" "$@"

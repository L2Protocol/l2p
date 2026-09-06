package systemcontracts

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

type UpgradeConfig struct {
	BeforeUpgrade upgradeHook
	AfterUpgrade  upgradeHook
	ContractAddr  common.Address
	CommitUrl     string
	Code          string
}

type Upgrade struct {
	UpgradeName string
	Configs     []*UpgradeConfig
}

type upgradeHook func(blockNumber *big.Int, contractAddr common.Address, statedb vm.StateDB) error

const (
	mainNet    = "Mainnet"
	defaultNet = "Default"
)

var (
	GenesisHash common.Hash
	// upgrade config
	ramanujanUpgrade = make(map[string]*Upgrade)

	eulerUpgrade = make(map[string]*Upgrade)

	planckUpgrade = make(map[string]*Upgrade)

	lubanUpgrade = make(map[string]*Upgrade)

	platoUpgrade = make(map[string]*Upgrade)

	keplerUpgrade = make(map[string]*Upgrade)

	feynmanUpgrade = make(map[string]*Upgrade)

	bohrUpgrade = make(map[string]*Upgrade)

	lorentzUpgrade = make(map[string]*Upgrade)

	maxwellUpgrade = make(map[string]*Upgrade)

	fermiUpgrade = make(map[string]*Upgrade)
)

func TryUpdateBuildInSystemContract(config *params.ChainConfig, blockNumber *big.Int, lastBlockTime uint64, blockTime uint64, statedb vm.StateDB, atBlockBegin bool) {
	if atBlockBegin {
		if !config.IsFeynman(blockNumber, lastBlockTime) {
			upgradeBuildInSystemContract(config, blockNumber, lastBlockTime, blockTime, statedb)
		}
		// HistoryStorageAddress is a special system contract in bsc, which can't be upgraded
		if config.IsOnPrague(blockNumber, lastBlockTime, blockTime) {
			statedb.SetCode(params.HistoryStorageAddress, params.HistoryStorageCode, tracing.CodeChangeSystemContractUpgrade)
			statedb.SetNonce(params.HistoryStorageAddress, 1, tracing.NonceChangeNewContract)
			log.Info("Set code for HistoryStorageAddress", "blockNumber", blockNumber.Int64(), "blockTime", blockTime)
		}
	} else {
		if config.IsFeynman(blockNumber, lastBlockTime) {
			upgradeBuildInSystemContract(config, blockNumber, lastBlockTime, blockTime, statedb)
		}
	}
}

func upgradeBuildInSystemContract(config *params.ChainConfig, blockNumber *big.Int, lastBlockTime uint64, blockTime uint64, statedb vm.StateDB) {
	if config == nil || blockNumber == nil || statedb == nil || reflect.ValueOf(statedb).IsNil() {
		return
	}

	var network string
	switch GenesisHash {
	/* Add mainnet genesis hash */
	case params.MainnetGenesisHash:
		network = mainNet
	default:
		network = defaultNet
	}

	logger := log.New("system-contract-upgrade", network)
	if config.IsOnRamanujan(blockNumber) {
		applySystemContractUpgrade(ramanujanUpgrade[network], blockNumber, statedb, logger)
	}

	if config.IsOnEuler(blockNumber) {
		applySystemContractUpgrade(eulerUpgrade[network], blockNumber, statedb, logger)
	}

	if config.IsOnPlanck(blockNumber) {
		applySystemContractUpgrade(planckUpgrade[network], blockNumber, statedb, logger)
	}

	if config.IsOnLuban(blockNumber) {
		applySystemContractUpgrade(lubanUpgrade[network], blockNumber, statedb, logger)
	}

	if config.IsOnPlato(blockNumber) {
		applySystemContractUpgrade(platoUpgrade[network], blockNumber, statedb, logger)
	}

	if config.IsOnShanghai(blockNumber, lastBlockTime, blockTime) {
		logger.Info("Empty upgrade config for shanghai", "height", blockNumber.String())
	}

	if config.IsOnKepler(blockNumber, lastBlockTime, blockTime) {
		applySystemContractUpgrade(keplerUpgrade[network], blockNumber, statedb, logger)
	}

	if config.IsOnFeynman(blockNumber, lastBlockTime, blockTime) {
		applySystemContractUpgrade(feynmanUpgrade[network], blockNumber, statedb, logger)
	}

	if config.IsOnBohr(blockNumber, lastBlockTime, blockTime) {
		applySystemContractUpgrade(bohrUpgrade[network], blockNumber, statedb, logger)
	}

	if config.IsOnLorentz(blockNumber, lastBlockTime, blockTime) {
		applySystemContractUpgrade(lorentzUpgrade[network], blockNumber, statedb, logger)
	}

	if config.IsOnMaxwell(blockNumber, lastBlockTime, blockTime) {
		applySystemContractUpgrade(maxwellUpgrade[network], blockNumber, statedb, logger)
	}

	if config.IsOnFermi(blockNumber, lastBlockTime, blockTime) {
		applySystemContractUpgrade(fermiUpgrade[network], blockNumber, statedb, logger)
	}

	/*
		apply other upgrades
	*/
}

func applySystemContractUpgrade(upgrade *Upgrade, blockNumber *big.Int, statedb vm.StateDB, logger log.Logger) {
	if upgrade == nil {
		logger.Info("Empty upgrade config", "height", blockNumber.String())
		return
	}

	logger.Info(fmt.Sprintf("Apply upgrade %s at height %d", upgrade.UpgradeName, blockNumber.Int64()))
	for _, cfg := range upgrade.Configs {
		logger.Info(fmt.Sprintf("Upgrade contract %s to commit %s", cfg.ContractAddr.String(), cfg.CommitUrl))

		if cfg.BeforeUpgrade != nil {
			err := cfg.BeforeUpgrade(blockNumber, cfg.ContractAddr, statedb)
			if err != nil {
				panic(fmt.Errorf("contract address: %s, execute beforeUpgrade error: %s", cfg.ContractAddr.String(), err.Error()))
			}
		}

		newContractCode, err := hex.DecodeString(strings.TrimSpace(cfg.Code))
		if err != nil {
			panic(fmt.Errorf("failed to decode new contract code: %s", err.Error()))
		}
		statedb.SetCode(cfg.ContractAddr, newContractCode, tracing.CodeChangeSystemContractUpgrade)

		if cfg.AfterUpgrade != nil {
			err := cfg.AfterUpgrade(blockNumber, cfg.ContractAddr, statedb)
			if err != nil {
				panic(fmt.Errorf("contract address: %s, execute afterUpgrade error: %s", cfg.ContractAddr.String(), err.Error()))
			}
		}
	}
}

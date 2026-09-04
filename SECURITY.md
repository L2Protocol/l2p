# Security Policy

## Supported Versions

Please see [Releases](https://github.com/L2Protocol/l2p/releases). We recommend using the [most recently released version](https://github.com/L2Protocol/l2p/releases/latest).

## Audit reports

This client is a fork of [bsc](https://github.com/bnb-chain/bsc), which is itself a fork of [go-ethereum](https://github.com/ethereum/go-ethereum). The audits below cover code that L2P still runs, and are published in the `docs/audits` folder of this repository.

| Scope    | Date     | Report Link                                                                                             |
| -------- | -------- | ------------------------------------------------------------------------------------------------------- |
| `geth`   | 20170425 | [pdf](docs/audits/2017-04-25_Geth-audit_Truesec.pdf)                                                     |
| `clef`   | 20180914 | [pdf](docs/audits/2018-09-14_Clef-audit_NCC.pdf)                                                         |
| `discv5` | 20191015 | [pdf](docs/audits/2019-10-15_Discv5_audit_LeastAuthority.pdf)                                            |
| `discv5` | 20200124 | [pdf](docs/audits/2020-01-24_DiscV5_audit_Cure53.pdf)                                                    |

No audit has been carried out on the L2P-specific changes.

## Reporting a Vulnerability

**Please do not file a public ticket** mentioning the vulnerability.

Report it privately through [GitHub Security Advisories](https://github.com/L2Protocol/l2p/security/advisories/new). Please include the affected version, a description of the impact, and steps to reproduce.

If the vulnerability originates in upstream code, we will coordinate disclosure with the relevant upstream project after assessing the impact on L2P.

// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package params

import "github.com/ethereum/go-ethereum/common"

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main BSC network.
var MainnetBootnodes = []string{
	"enode://ce977fa4e9662712ee3d8ebdd3ecc58c85f5eb0ccf453d7a17b97b5b12084f573f137998e2c9b696148a86613eb64c99b8cf24a1c841e45dee46a5dbcb3fde6f@46.62.251.87:31398",
	"enode://5906da28b23b6020c28e9ddc22cfc7a3b2f294c0f6164beb74e3825a18d4ab9c2f4394f345de191d272723a61f647d448d37c81aada0a84b96f5a1fc09b050ec@128.140.59.159:31398",
	"enode://fa7ab5cf291659f3615665a27cd623255a9d6c92edcf50395940ae86be1f7414310f0ecd5b5c22b09e56fc1f49b8e79d35f2bf5165b963aacb5619e0a6032dbe@157.180.90.222:31398",
	"enode://79b5e81a8162616b81fa076c501d3380fac47807eaa5a4c5068fdd8f531d7652fbb940f88d8d1df4e67c8eb0bdcb91c9b2efda468749116f31770ea700abff21@46.224.98.68:31398",
}

var V5Bootnodes = []string{
	"enr:-Ki4QLCHmVCpIBMcsq8vFe1a3UKVzdyrKi-EqkC5rgf53kxYfbGInAFvXhVB4it4d1-MdCSH-UI2pIX_IoZ2c3OduX6GAZrFPWZOg2JzY8CDZXRox8aEY_-_-AGCaWSCdjSCaXCELj77V4lzZWNwMjU2azGhA86Xf6TpZicS7j2OvdPsxYyF9esMz0U9ehe5e1sSCE9XhHNuYXDAg3RjcIJ6poN1ZHCCeqY",
	"enr:-Ki4QFrpxU9iCpGgnbxhSzpsO9g_jTVJKmknAnkHc6zzZ1PLOHvj6f47VyYnffumuy8BArmx8th2kThucciIDzpff2yGAZrFQBxNg2JzY8CDZXRox8aEY_-_-AGCaWSCdjSCaXCEgIw7n4lzZWNwMjU2azGhAlkG2iiyO2Agwo6d3CLPx6Oy8pTA9hZL63TjgloY1KuchHNuYXDAg3RjcIJ6poN1ZHCCeqY",
	"enr:-Ki4QGt-Va0XjIsuBBsYuhYC8CBvQW34FjGeacw-f4k_sw3RebXtyw8hRrxLmWI8Pnp5FW0ygJRZQGIC_4Hf-RYWlNCGAZrz_5Lmg2JzY8CDZXRox8aEKmaNL4CCaWSCdjSCaXCEnbRa3olzZWNwMjU2azGhAvp6tc8pFlnzYVZlonzWIyVanWyS7c9QOVlAroa-H3QUhHNuYXDAg3RjcIJ6poN1ZHCCeqY",
	"enr:-Ki4QNLwTiCVafsORQxl2JBYXPW6RZZvJ6iNPBksPHjYsvgIU9LZ_t-pA_xmGY4DLZnj8kEBZOXH7Z4rTjWsjUSLAmOGAZr0BTGBg2JzY8CDZXRox8aEKmaNL4CCaWSCdjSCaXCELuBiRIlzZWNwMjU2azGhA3m16BqBYmFrgfoHbFAdM4D6xHgH6qWkxQaP3Y9THXZShHNuYXDAg3RjcIJ6poN1ZHCCeqY",
}

const dnsPrefix = "enrtree://AP6MZV3RNCCTYM3EEM5YRAMWT4L53A4H4YTHGDH4HD2O3AOR4N6LK@"

// KnownDNSNetwork returns the address of a public DNS-based node list for the given
// genesis hash and protocol. See https://github.com/ethereum/discv4-dns-lists for more
// information.
func KnownDNSNetwork(genesis common.Hash, protocol string) string {
	var net string
	switch genesis {
	case L2PGenesisHash:
		net = "mainnet"
	default:
		return ""
	}
	return dnsPrefix + protocol + "." + net + ".l2pdisco.net"
}

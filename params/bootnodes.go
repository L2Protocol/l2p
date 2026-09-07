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
// the L2P mainnet.
var MainnetBootnodes = []string{
	"enode://fa7ab5cf291659f3615665a27cd623255a9d6c92edcf50395940ae86be1f7414310f0ecd5b5c22b09e56fc1f49b8e79d35f2bf5165b963aacb5619e0a6032dbe@212.95.41.176:31398",
	"enode://79b5e81a8162616b81fa076c501d3380fac47807eaa5a4c5068fdd8f531d7652fbb940f88d8d1df4e67c8eb0bdcb91c9b2efda468749116f31770ea700abff21@212.95.41.175:31398",
	"enode://cdebca373c4e5bb737182abcf0e536ac40485be777f19a85e0d900b1f3ebf9b61efdb6d4addfcf1d4d660aaacccca42b3ba330e31f569aa3f539b8e1993f2eb0@212.95.41.173:31398",
}

var V5Bootnodes = []string{
	"enr:-Iu4QNGXIXHIZazZBSnNqQBNn3GLydTHUtPptuaCC9cgEf1UAPbfk4J95hf25CV1AcX7GkdIgj1FXPv4gojSC-OTAd-AgmlkgnY0gmlwhNRfKbCJc2VjcDI1NmsxoQL6erXPKRZZ82FWZaJ81iMlWp1sku3PUDlZQK6Gvh90FIN0Y3CCeqaDdWRwgnqm",
	"enr:-Iu4QCnPl_YmwknpIOOYpSTxtTIEjFcYbb0CKekXmSBH61SCTErLBryOmVpsqrT-FZenlnVqR1NbC9w3DMobeyOwcjSAgmlkgnY0gmlwhNRfKa-Jc2VjcDI1NmsxoQN5tegagWJha4H6B2xQHTOA-sR4B-qlpMUGj92PUx12UoN0Y3CCeqaDdWRwgnqm",
	"enr:-Iu4QB_rRS7mbVbtQb1GxP3RYsimIfKk6A5iyvBdIXPCmKiyV6fClUwLhvo2FR4bzoNZ44lYQcf-NSEF7AuTOu7OcZSAgmlkgnY0gmlwhNRfKa2Jc2VjcDI1NmsxoQLN68o3PE5btzcYKrzw5TasQEhb53fxmoXg2QCx8-v5toN0Y3CCeqaDdWRwgnqm",
}

const dnsPrefix = "enrtree://AP6MZV3RNCCTYM3EEM5YRAMWT4L53A4H4YTHGDH4HD2O3AOR4N6LK@"

// KnownDNSNetwork returns the address of a public DNS-based node list for the given
// genesis hash and protocol. See https://github.com/ethereum/discv4-dns-lists for more
// information.
func KnownDNSNetwork(genesis common.Hash, protocol string) string {
	var net string
	switch genesis {
	case MainnetGenesisHash:
		net = "mainnet"
	default:
		return ""
	}
	return dnsPrefix + protocol + "." + net + ".l2pdisco.net"
}

////////////////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 The mjoy-go Authors.
//
// The mjoy-go is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// @File: nodeevent_string.go
// @Date: 2018/05/08 15:18:08
////////////////////////////////////////////////////////////////////////////////

// Code generated by "stringer -type=nodeEvent"; DO NOT EDIT.

package discv5

import "strconv"

const (
	_nodeEvent_name_0 = "invalidEventpingPacketpongPacketfindnodePacketneighborsPacketfindnodeHashPackettopicRegisterPackettopicQueryPackettopicNodesPacket"
	_nodeEvent_name_1 = "pongTimeoutpingTimeoutneighboursTimeout"
)

var (
	_nodeEvent_index_0 = [...]uint8{0, 12, 22, 32, 46, 61, 79, 98, 114, 130}
	_nodeEvent_index_1 = [...]uint8{0, 11, 22, 39}
)

func (i nodeEvent) String() string {
	switch {
	case 0 <= i && i <= 8:
		return _nodeEvent_name_0[_nodeEvent_index_0[i]:_nodeEvent_index_0[i+1]]
	case 265 <= i && i <= 267:
		i -= 265
		return _nodeEvent_name_1[_nodeEvent_index_1[i]:_nodeEvent_index_1[i+1]]
	default:
		return "nodeEvent(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}

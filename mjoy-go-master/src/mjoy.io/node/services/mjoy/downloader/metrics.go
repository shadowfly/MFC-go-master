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
// @File: metrics.go
// @Date: 2018/05/08 18:02:08
////////////////////////////////////////////////////////////////////////////////

// Contains the metrics collected by the downloader.

package downloader

import (
	"mjoy.io/utils/metrics"
)

var (
	headerInMeter      = metrics.GetOrRegisterMeter("mjoy/downloader/headers/in",metrics.DefaultRegistry)
	headerReqTimer     = metrics.NewRegisteredTimer("mjoy/downloader/headers/req",metrics.DefaultRegistry)
	headerDropMeter    = metrics.GetOrRegisterMeter("mjoy/downloader/headers/drop",metrics.DefaultRegistry)
	headerTimeoutMeter = metrics.GetOrRegisterMeter("mjoy/downloader/headers/timeout",metrics.DefaultRegistry)

	bodyInMeter      = metrics.GetOrRegisterMeter("mjoy/downloader/bodies/in",metrics.DefaultRegistry)
	bodyReqTimer     = metrics.NewRegisteredTimer("mjoy/downloader/bodies/req",metrics.DefaultRegistry)
	bodyDropMeter    = metrics.GetOrRegisterMeter("mjoy/downloader/bodies/drop",metrics.DefaultRegistry)
	bodyTimeoutMeter = metrics.GetOrRegisterMeter("mjoy/downloader/bodies/timeout",metrics.DefaultRegistry)

	receiptInMeter      = metrics.GetOrRegisterMeter("mjoy/downloader/receipts/in",metrics.DefaultRegistry)
	receiptReqTimer     = metrics.NewRegisteredTimer("mjoy/downloader/receipts/req",metrics.DefaultRegistry)
	receiptDropMeter    = metrics.GetOrRegisterMeter("mjoy/downloader/receipts/drop",metrics.DefaultRegistry)
	receiptTimeoutMeter = metrics.GetOrRegisterMeter("mjoy/downloader/receipts/timeout",metrics.DefaultRegistry)

	stateInMeter   = metrics.GetOrRegisterMeter("mjoy/downloader/states/in",metrics.DefaultRegistry)
	stateDropMeter = metrics.GetOrRegisterMeter("mjoy/downloader/states/drop",metrics.DefaultRegistry)

)

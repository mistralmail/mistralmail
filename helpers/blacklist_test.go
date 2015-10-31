package helpers

import (
	. "github.com/smartystreets/goconvey/convey"
	"sort"
	"testing"
)

func TestParseAddress(t *testing.T) {

	Convey("Testing CheckIp()", t, func() {
		ns := Nixspam{
			IpList: []string{
				"91.210.25.14",
				"175.156.192.59",
				"79.13.109.29",
				"23.250.116.32",
				"24.102.174.37",
				"2abe355e50e33850ec12b72ce85659cb",
				"113.203.235.30",
				"212.124.28.9",
				"75.97.65.86",
				"cbd08af21102f448c3f7d2b99c107383",
				"182.172.111.91",
				"49387a67d9f27c7279d29369a863ed08",
			},
		}

		sort.Strings(ns.IpList)

		So(ns.CheckIp("91.210.25.14"), ShouldEqual, true)
		So(ns.CheckIp("23.250.116.32"), ShouldEqual, true)
		So(ns.CheckIp("2abe355e50e33850ec12b72ce85659cb"), ShouldEqual, true)
		So(ns.CheckIp("49387a67d9f27c7279d29369a863ed08"), ShouldEqual, true)

		So(ns.CheckIp("192.168.0.10"), ShouldEqual, false)
		So(ns.CheckIp("2a0267e0000000000000000000000010"), ShouldEqual, false)
	})

	// Just some manual test I wrote with the current db
	// note that this db can (and will) change
	//Convey("Testing downloading the db", t, func() {
	//	ns, err := NewNixspam()
	//	So(err, ShouldEqual, nil)
	//	So(ns.CheckIp("91.210.25.14"), ShouldEqual, true)
	//	So(ns.CheckIp("23.250.116.32"), ShouldEqual, true)
	//	So(ns.CheckIp("cbd08af21102f448c3f7d2b99c107383"), ShouldEqual, true)
	//	So(ns.CheckIp("2abe355e50e33850ec12b72ce85659cb"), ShouldEqual, true)
	//})

}

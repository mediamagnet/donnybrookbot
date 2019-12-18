// Copyright (c) 2015, Sgt. Kabukiman | MIT licensed

package srapi

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRegions(t *testing.T) {
	countRequests = true

	iQue := "mol4z19n"
	pal := "e6lxy1dz"

	Convey("Fetching regions by valid IDs", t, func() {
		region, err := RegionByID(iQue)
		So(err, ShouldBeNil)
		So(region.ID, ShouldEqual, iQue)
		So(region.Name, ShouldEqual, "CHN / iQue")
		So(region.Links, ShouldNotBeEmpty)
	})

	Convey("Fetching regions by invalid IDs", t, func() {
		region, err := RegionByID("i_do_not_exist")
		So(err, ShouldNotBeNil)
		So(region, ShouldBeNil)
	})

	Convey("Fetching multiple regions", t, func() {
		regions, err := Regions(nil, &Cursor{0, 1})
		So(err, ShouldBeNil)
		So(regions.Pagination.Offset, ShouldEqual, 0)
		So(regions.Pagination.Max, ShouldEqual, 1)

		num := 0

		// read a few pages, 7 is arbitrary
		regions.Walk(func(r *Region) bool {
			So(r.ID, ShouldNotBeBlank)

			num++
			return num < 7
		})
	})

	Convey("Fetching runs of a region", t, func() {
		region, err := RegionByID(pal)
		So(err, ShouldBeNil)

		runs, err := region.Runs(nil, nil, NoEmbeds)
		So(err, ShouldBeNil)

		firstID := ""

		Convey("first page of runs should be fine", func() {
			So(runs.Data, ShouldNotBeEmpty)
			So(runs.Pagination.Offset, ShouldEqual, 0)

			firstID = runs.Data[0].ID
		})

		runs, err = region.Runs(nil, &Sorting{Direction: Descending}, NoEmbeds)
		So(err, ShouldBeNil)

		Convey("sorting order should be taken into account", func() {
			So(runs.Data, ShouldNotBeEmpty)
			So(runs.Pagination.Offset, ShouldEqual, 0)
			So(firstID, ShouldNotEqual, runs.Data[0].ID)
		})
	})

	Convey("Fetching games of a region", t, func() {
		region, err := RegionByID(pal)
		So(err, ShouldBeNil)

		games, err := region.Games(nil, nil, NoEmbeds)
		So(err, ShouldBeNil)

		firstID := ""

		Convey("first page of games should be fine", func() {
			So(games.Data, ShouldNotBeEmpty)
			So(games.Pagination.Offset, ShouldEqual, 0)

			firstID = games.Data[0].ID
		})

		games, err = region.Games(nil, &Sorting{Direction: Descending}, NoEmbeds)
		So(err, ShouldBeNil)

		Convey("sorting order should be taken into account", func() {
			So(games.Data, ShouldNotBeEmpty)
			So(games.Pagination.Offset, ShouldEqual, 0)
			So(firstID, ShouldNotEqual, games.Data[0].ID)
		})
	})
}

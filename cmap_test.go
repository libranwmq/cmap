package cmap

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
)

func TestUnit_ConcurrentMap(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	opts := []ConcurrentMapOption[String, int]{
		WithConcurrentMapSlotNum[String, int](16),
	}
	scm := New[String, int](opts...)

	convey.Convey("TestUnit_ConcurrentMap", t, func() {
		convey.Convey("ConcurrentMap test-01", func() {
			cnt := scm.Count()
			convey.So(cnt, convey.ShouldEqual, 0)
			num := scm.GetSlotNum()
			convey.So(num, convey.ShouldEqual, 16)
			isEmpty := scm.IsEmpty()
			convey.So(isEmpty, convey.ShouldBeTrue)

			scm.Set("key1", 1)
			cnt = scm.Count()
			convey.So(cnt, convey.ShouldEqual, 1)

			ok := scm.SetIfAbsent("key1", 2)
			convey.So(ok, convey.ShouldBeFalse)
			isEmpty = scm.IsEmpty()
			convey.So(isEmpty, convey.ShouldBeFalse)

			data := map[String]int{
				"key2": 2,
				"key3": 3,
			}
			scm.MSet(data)

			cnt = scm.Count()
			convey.So(cnt, convey.ShouldEqual, 3)

			v, ok := scm.Get("key2")
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(v, convey.ShouldEqual, 2)

			v, ok = scm.Pop("key3")
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(v, convey.ShouldEqual, 3)

			ok = scm.Has("key1")
			convey.So(ok, convey.ShouldBeTrue)

			scm.Remove("key1")
			ok = scm.Has("key1")
			convey.So(ok, convey.ShouldBeFalse)

			m := scm.Items()
			convey.So(len(m), convey.ShouldEqual, 1)
			keys := scm.Keys()
			convey.So(len(keys), convey.ShouldEqual, 1)

			for kvp := range scm.Iter() {
				convey.So(kvp.Key, convey.ShouldEqual, String("key2"))
				convey.So(kvp.Value, convey.ShouldEqual, 2)
			}

			scm.Clear()
		})
	})
}

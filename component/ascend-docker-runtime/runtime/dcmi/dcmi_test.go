package dcmi

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestGetChipInfo(t *testing.T) {
	convey.Convey("Test get chip function", t, func() {
		var cardId int32 = 0
		var devId int32 = 0
		w := NpuWorker{}
		convey.Convey("01-not valid id, should return error", func() {
			gomonkey.ApplyFunc(isValidCardIDAndDeviceID, func(a, b int32) bool {
				return false
			})
			_, err := w.GetChipInfo(cardId, devId)
			convey.ShouldBeError(err)
		})
		convey.Convey("02-get no chip info, should return nil", func() {
			gomonkey.ApplyFunc(isValidCardIDAndDeviceID, func(a, b int32) bool {
				return true
			})
			chip, _ := w.GetChipInfo(cardId, devId)
			convey.ShouldBeNil(chip)
		})
	})
}

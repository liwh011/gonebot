package event

import "testing"

func Test_GetEventName(t *testing.T) {
	testData := []struct {
		event  T_Event
		expect string
	}{
		{
			GroupAdminNoticeEvent{
				NoticeEvent: NoticeEvent{
					Event{
						PostType: "notice",
					},
					"group",
				},
				SubType: "set",
			},
			"notice.group.set",
		},
		{
			GroupUploadNoticeEvent{
				NoticeEvent: NoticeEvent{
					Event{
						PostType: "notice",
					},
					"group_upload",
				},
			},
			"notice.group_upload",
		},
	}

	for _, data := range testData {
		if got := GetEventName(&data.event); got != data.expect {
			t.Errorf("GetEventName(%#v) = %#v; expect %#v", data.event, got, data.expect)
		}
	}
}

package uchat

import (
	"testing"
)

var testClient *UchatClient
var testRobot map[string]string
var testChatRooms []map[string]string

func init() {
	testClient = NewClient("201705311010001", "123123")
}

func Test_001_RobotList(t *testing.T) {
	rst, err := testClient.RobotList()
	if err != nil {
		t.Error(err)
	} else {
		t.Log(rst)
		testRobot = rst[0]
	}
}

func Test_002_ApplyCodeList(t *testing.T) {
	ctx := make(map[string]string, 0)
	ctx["vcRobotSerialNo"] = testRobot["vcSerialNo"]
	ctx["nType"] = "10"
	ctx["vcChatRoomSerialNo"] = ""
	ctx["nCodeCount"] = "2"
	ctx["nAddMinute"] = "1"
	ctx["nIsNotify"] = "0"
	ctx["vcNotifyContent"] = ""
	rst, err := testClient.ApplyCodeList(ctx)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(rst)
	}
}

func Test_003_ChatRoomList(t *testing.T) {
	ctx := make(map[string]string, 0)
	ctx["vcRobotSerialNo"] = testRobot["vcSerialNo"]
	rst, err := testClient.ChatRoomList(ctx)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(rst)
		testChatRooms = rst
	}
}

func Test_004_ChatRoomUserInfo(t *testing.T) {
	for _, v := range testChatRooms {
		ctx := make(map[string]string, 0)
		ctx["vcChatRoomSerialNo"] = v["vcChatRoomSerialNo"]
		//ctx["vcChatRoomSerialNo"] = "201706141050000018"
		err := testClient.ChatRoomUserInfo(ctx)
		if err != nil {
			t.Error(err)
		} else {
			t.Log("success")
		}
	}
}

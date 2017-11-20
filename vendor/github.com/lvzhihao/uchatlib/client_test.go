package uchatlib

import (
	"flag"
	"testing"

	"github.com/lvzhihao/goutils"
)

var testClient *UchatClient
var testRobot map[string]string
var testChatRooms []map[string]string

var testMarchantNo string
var testMmarchantSecret string

func init() {
	flag.StringVar(&testMarchantNo, "no", "", "marchantNo")
	flag.StringVar(&testMmarchantSecret, "secret", "", "marchantSecret")
	flag.Parse()
	testClient = NewClient(testMarchantNo, testMmarchantSecret)
}

func Test_001_RobotList(t *testing.T) {
	rst, err := testClient.RobotList()
	if err != nil {
		t.Error(err)
	} else {
		t.Log(rst)
		var begin int32 = 0
		//find most chats to testing
		for _, v := range rst {
			count := goutils.ToInt32(v["nChatRoomCount"])
			if count > begin {
				begin = count
				testRobot = v
			}
		}
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

func Test_004_ChatRoomStatus(t *testing.T) {
	for k, v := range testChatRooms {
		ctx := make(map[string]string, 0)
		ctx["vcChatRoomSerialNo"] = v["vcChatRoomSerialNo"]
		data, err := testClient.ChatRoomStatus(ctx)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(data)
			testChatRooms[k]["nRobotInStatus"] = data["nRobotInStatus"]
			testChatRooms[k]["nRobotStatus"] = data["nRobotStatus"]
			testChatRooms[k]["nStatus"] = data["nStatus"]
			t.Log("success")
		}
	}
}

func Test_005_ChatRoomUserInfo(t *testing.T) {
	for _, v := range testChatRooms {
		if v["nRobotInStatus"] == "1" {
			continue
		}
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

func Test_006_ChatRoomCloseGetMessage(t *testing.T) {
	for _, v := range testChatRooms {
		ctx := make(map[string]string, 0)
		ctx["vcChatRoomSerialNo"] = v["vcChatRoomSerialNo"]
		err := testClient.ChatRoomCloseGetMessages(ctx)
		if err != nil {
			t.Error(err)
		} else {
			t.Log("success")
		}
	}
}

func Test_007_ChatRoomOpenGetMessage(t *testing.T) {
	for _, v := range testChatRooms {
		ctx := make(map[string]string, 0)
		ctx["vcChatRoomSerialNo"] = v["vcChatRoomSerialNo"]
		err := testClient.ChatRoomOpenGetMessages(ctx)
		if err != nil {
			t.Error(err)
		} else {
			t.Log("success")
		}
	}
}

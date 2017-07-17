package uchat

//小U机器回调队列配置
var (
	ReceiveMQMemberList          = "uchat.member.list"           //用户列表
	ReceiveMQMemberJoin          = "uchat.member.join"           //用户入群
	ReceiveMQMemberQuit          = "uchat.member.quit"           //用户群群
	ReceiveMQMemberMessageSum    = "uchat.member.message.sum"    //用户发言总数
	ReceiveMQRobotChatList       = "uchat.robot.chat.list"       //设备所开群列表
	ReceiveMQRobotJoinChat       = "uchat.robot.chat.join"       //设备开群信息
	ReceiveMQRobotPrivateMessage = "uchat.robot.message.private" //设备私聊
	ReceiveMQChatKeyword         = "uchat.chat.keyword"          //群关键字
	ReceiveMQChatCreate          = "uchat.chat.create"           //建群
	ReceiveMQChatMessage         = "uchat.chat.message"          //群聊天记录
	ReceiveMQChatRedpack         = "uchat.chat.redpack"          //群红包记录
)

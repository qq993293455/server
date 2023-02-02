package module

import (
	Arena "coin-server/load-test/module/arena"
)

func Init() {
	// 不能注释user.New，否则心跳会断开
	//Register(user.New, 1000)
	// Register(mail.New)
	// Register(bag.New, 1)
	/*
		Register(im.New)*/
	// Register(achievement.New, 1000)
	// Register(match.NewOwner, 500)
	//Register(match_joiner.NewJoiner, 1000)
	//Register(task.New)
	//Register(maintask.New)
	// Register(guild.New, 50000)
	// Register(friend.New, 1000)
	//Register(shop.New, 1000)
	//Register(divination.New, 50000)
	Register(Arena.New, 1000)
}

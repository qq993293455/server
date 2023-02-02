package enum

// 新增key的时候一定要将key添加到AllRankKey中，否则服务启动的时候不会把数据加载进来
const GuildRankKey = "guild_rank_id"

func AllRankKey() []string {
	return []string{
		GuildRankKey,
	}
}

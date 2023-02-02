package rule_model

func Parsing(data *Data, custom CustomParse) {
	h.version = data.GetVersion()
	r.SetVersion(h.version)
	r.SetCustom(custom)

	ParsePvPRankReward(data)
	ParsePvPRobot(data)
	ParseAchievement(data)
	ParseAchievementList(data)
	ParseAchievementTotal(data)
	ParseActionfusion(data)
	ParseActivity(data)
	ParseActivity0yuanpurchase(data)
	ParseActivityXdaygoal(data)
	ParseActivityXdaygoalScorereward(data)
	ParseActivityAccumulatedrecharge(data)
	ParseActivityCircular(data)
	ParseActivityDailygiftBuy(data)
	ParseActivityFirstpay(data)
	ParseActivityGrowthfund(data)
	ParseActivityLimitedtimePackage(data)
	ParseActivityLimitedtimePackagePay(data)
	ParseActivityLoginReward(data)
	ParseActivityMonthlycard(data)
	ParseActivityPassesAwards(data)
	ParseActivityReward(data)
	ParseActivityStellargemShopmall(data)
	ParseActivityWeekly(data)
	ParseActivityWeeklyCastsword(data)
	ParseActivityWeeklyChallenge(data)
	ParseActivityWeeklyExchange(data)
	ParseActivityWeeklyGift(data)
	ParseActivityWeeklyRank(data)
	ParseAnecdotes(data)
	ParseAnecdotesClassText(data)
	ParseAnecdotesGame1(data)
	ParseAnecdotesGame1Option(data)
	ParseAnecdotesGame11(data)
	ParseAnecdotesGame2(data)
	ParseAnecdotesGame3(data)
	ParseAnecdotesGame3Option(data)
	ParseAnecdotesGame4(data)
	ParseAnecdotesGame4Option(data)
	ParseAnecdotesGame5(data)
	ParseAnecdotesGame5Option(data)
	ParseAnecdotesGame7(data)
	ParseAnecdotesGame7Option(data)
	ParseAnecdotesGame8(data)
	ParseAnecdotesGame8Option(data)
	ParseAnecdotesGames(data)
	ParseAnecdotesReward(data)
	ParseAnnouncement(data)
	ParseAtlas(data)
	ParseAtlasItem(data)
	ParseAttr(data)
	ParseAttrTrans(data)
	ParseAutoWindow(data)
	ParseBag(data)
	ParseBanner(data)
	ParseBeStronger(data)
	ParseBegining(data)
	ParseBiography(data)
	ParseBossHall(data)
	ParseBuff(data)
	ParseBuffEffect(data)
	ParseBuffEffectShow(data)
	ParseBuffLogic(data)
	ParseCharge(data)
	ParseClassAh(data)
	ParseCombatBattle(data)
	ParseDivination(data)
	ParseDrop(data)
	ParseDropMini(data)
	ParseDropLists(data)
	ParseDropListsMini(data)
	ParseDubbing(data)
	ParseDungeon(data)
	ParseEffect(data)
	ParseEmoji(data)
	ParseEnchantments(data)
	ParseEntrySkill(data)
	ParseEquip(data)
	ParseEquipComplete(data)
	ParseEquipEntry(data)
	ParseEquipLight(data)
	ParseEquipQuality(data)
	ParseEquipRefine(data)
	ParseEquipResonance(data)
	ParseEquipSlot(data)
	ParseEquipStar(data)
	ParseEquipType(data)
	ParseEquipment(data)
	ParseExchange(data)
	ParseExchangeItem(data)
	ParseExpSkip(data)
	ParseExpedition(data)
	ParseExpeditionReward(data)
	ParseExpeditionQuantity(data)
	ParseFailureToStronger(data)
	ParseFashion(data)
	ParseFashionActivate(data)
	ParseFightingCapacity(data)
	ParseFlynum(data)
	ParseForgeFixedRecipe(data)
	ParseForgeLevel(data)
	ParseForgeSupplement(data)
	ParseGacha(data)
	ParseGachalist(data)
	ParseGift(data)
	ParseGiftItem(data)
	ParseGuideProcess(data)
	ParseGuideStep(data)
	ParseGuild(data)
	ParseGuildJournal(data)
	ParseGuildContend(data)
	ParseGuildContendbuild(data)
	ParseGuildIntegral(data)
	ParseGuildBlessing(data)
	ParseGuildBoss(data)
	ParseGuildFunction(data)
	ParseGuildGvghelp(data)
	ParseGuildPosition(data)
	ParseGuildSign(data)
	ParseGuilde(data)
	ParseHeadSculpture(data)
	ParseHeadSpeak(data)
	ParseHeadSpeakWord(data)
	ParseHeadTalkBubble(data)
	ParseHero(data)
	ParseHeroDeviationEntry(data)
	ParseHomeAvatar(data)
	ParseImageDisplayLocation(data)
	ParseImgLanguage(data)
	ParseInhibitAtk(data)
	ParseItem(data)
	ParseItemClass(data)
	ParseItemTypeIcon(data)
	ParseJourneyChest(data)
	ParseJourneyList(data)
	ParseKeyValue(data)
	ParseLanguageBackend(data)
	ParseLooptask(data)
	ParseLoopTaskStageReward(data)
	ParseMail(data)
	ParseMainCityMap(data)
	ParseMainCityMapBuilding(data)
	ParseMainCityMapNpcPath(data)
	ParseMainTask(data)
	ParseMainTaskChapter(data)
	ParseMainTaskChapterReward(data)
	ParseMainTaskChapterTarget(data)
	ParseMapBuilding(data)
	ParseMapNpc(data)
	ParseMapNpcWithMainStory(data)
	ParseMapRelation(data)
	ParseMapScene(data)
	ParseMapSelect(data)
	ParseMapTransmit(data)
	ParseMechanics(data)
	ParseMedicament(data)
	ParseMonster(data)
	ParseMonsterAttr(data)
	ParseMonsterGroup(data)
	ParseNotice(data)
	ParseNpcChallengeDungeon(data)
	ParseNpcDialogue(data)
	ParseDialogueWord(data)
	ParseDialogueOpt(data)
	ParseNpcTask(data)
	ParseOptionalBox(data)
	ParsePersonalBossBuff(data)
	ParsePersonalBossHelpPointsReward(data)
	ParsePersonalBossHelperCards(data)
	ParsePersonalBossLibrary(data)
	ParsePersonalBossNumberFloor(data)
	ParsePersonalBossLv(data)
	ParsePersonalBossPassingReward(data)
	ParsePlaneDungeon(data)
	ParsePopup(data)
	ParseQuality(data)
	ParseRecycleType(data)
	ParseRelics(data)
	ParseRelicsFunctionAttr(data)
	ParseRelicsLv(data)
	ParseRelicsLvQuality(data)
	ParseRelicsSkill(data)
	ParseRelicsSkilltype(data)
	ParseRobot(data)
	ParseRobotName(data)
	ParseRoguelikeArtifact1(data)
	ParseRoguelikeArtifact2(data)
	ParseRoguelikeArtifact3(data)
	ParseRoguelikeDungeon(data)
	ParseRoguelikeDungeonRoom(data)
	ParseRoguelikeEntry(data)
	ParseRoleChangemodel(data)
	ParseRoleLv(data)
	ParseRoleLvTitle(data)
	ParseRoleReachDungeon(data)
	ParseRoleReachItem(data)
	ParseRoleSkill(data)
	ParseSkillLv(data)
	ParseRowHero(data)
	ParseSample(data)
	ParseSampleChild1(data)
	ParseSampleChild2(data)
	ParseSensitiveDict1(data)
	ParseSensitiveDict2(data)
	ParseShieldedWord(data)
	ParseShopgoodslist(data)
	ParseGoodslistLv(data)
	ParseSideSysEntry(data)
	ParseSkill(data)
	ParseSkillPos(data)
	ParseSkillstone(data)
	ParseSoulContract(data)
	ParseSpecialResBar(data)
	ParseStoryIllustration(data)
	ParseStringCn(data)
	ParseStringEn(data)
	ParseSummoned(data)
	ParseSystem(data)
	ParseSystemUnlockCondition(data)
	ParseTalent(data)
	ParseTalentAttr(data)
	ParseTalentplate(data)
	ParseTargetTask(data)
	ParseTargetTaskStage(data)
	ParseTaskType(data)
	ParseTempBag(data)
	ParseTestBattle(data)
	ParseTimeline(data)
	ParseTowerDefault(data)
	ParseVerifyLanguage(data)

	ParseCustom()
}

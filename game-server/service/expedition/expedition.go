package expedition

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/logger"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/timer"
	wr "coin-server/common/utils/weightedrand"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/module"
	"coin-server/game-server/service/expedition/dao"
	"coin-server/game-server/service/expedition/rule"
	rule_model "coin-server/rule/rule-model"

	"github.com/rs/xid"
	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func NewExpeditionService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	module *module.Module,
	log *logger.Logger,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		log:        log,
		Module:     module,
	}
	return s
}

func (svc *Service) Router() {
	svc.svc.RegisterFunc("获取远征信息", svc.Info)
	svc.svc.RegisterFunc("派遣", svc.Dispatch)
	svc.svc.RegisterFunc("领取任务完成奖励", svc.GetDoneReward)
	svc.svc.RegisterFunc("花费钻石立即完成任务", svc.DoneImmediately)
	svc.svc.RegisterFunc("刷新任务", svc.RefreshTask)

	eventlocal.SubscribeEventLocal(svc.HandleTaskChange)
	eventlocal.SubscribeEventLocal(svc.HandleExtraSkillTypTotalUpdate)
}

func (svc *Service) Info(ctx *ctx.Context, _ *servicepb.Expedition_ExpeditionInfoRequest) (*servicepb.Expedition_ExpeditionInfoResponse, *errmsg.ErrMsg) {
	e, ok, err := dao.Get(ctx)
	if err != nil {
		return nil, err
	}
	// 玩家首次进入该系统，初始化任务栏
	if !ok {
		slotCount, err := svc.getSlotCount(ctx)
		if err != nil {
			return nil, err
		}
		e.NormalSlot = slotCount
	}
	save1, err := svc.genTask(ctx, e, false, false)
	if err != nil {
		return nil, err
	}
	save2 := svc.handleFreeRefresh(ctx, e)
	count, recoverTime, save3, err := svc.handleExecutionRecover(ctx, e)
	if err != nil {
		return nil, err
	}

	if save1 || save2 || save3 {
		dao.Save(ctx, e)
	}
	return &servicepb.Expedition_ExpeditionInfoResponse{
		MustCount:             svc.getMustCount(ctx, e),
		ExecutionRecoverCount: count,
		NextRecoverTime:       recoverTime,
		Task:                  e.Task,
		ExtraSlot:             0, // TODO 目前还没有这块的设计
		FreeRefreshCount:      svc.getFreeRefreshCount(ctx, e),
		MaxExecution:          svc.getMaxExecution(ctx, e.Execution),
	}, nil
}

func (svc *Service) Dispatch(ctx *ctx.Context, req *servicepb.Expedition_ExpeditionDispatchRequest) (*servicepb.Expedition_ExpeditionDispatchResponse, *errmsg.ErrMsg) {
	e, ok, err := dao.Get(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrExpeditionTaskNotExist()
	}
	task, ok := e.Task[req.TaskId]
	if !ok {
		return nil, errmsg.NewErrExpeditionTaskNotExist()
	}
	cfg, ok := rule.GetExpeditionById(ctx, task.ConfigId)
	if !ok {
		return nil, errmsg.NewErrExpeditionTaskNotExist()
	}
	if len(cfg.LimitCondition) == 3 {
		taskType := models.TaskType(cfg.LimitCondition[0])
		condition := cfg.LimitCondition[1]
		need := cfg.LimitCondition[2]
		data, err := svc.GetCounterByType(ctx, taskType)
		if err != nil {
			return nil, err
		}
		if data[condition] < need {
			return nil, errmsg.NewErrLimitConditionNotEnough()
		}
	}

	if _, _, _, err = svc.handleExecutionRecover(ctx, e); err != nil {
		return nil, err
	}

	if err := svc.SubManyItem(ctx, ctx.RoleId, cfg.Cost); err != nil {
		return nil, err
	}
	task.Ongoing = true
	task.DoneTime = timer.StartTime(ctx.StartTime).Add(time.Duration(cfg.Duration) * time.Minute).Unix()
	e.Task[req.TaskId] = task

	svc.handleFreeRefresh(ctx, e)

	dao.Save(ctx, e)

	tasks := map[models.TaskType]*models.TaskUpdate{
		models.TaskType_TaskExpeditionNumAcc: {
			Typ:     values.Integer(models.TaskType_TaskExpeditionNumAcc),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskExpeditionNumA: {
			Typ:     values.Integer(models.TaskType_TaskExpeditionNumA),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
	}
	svc.UpdateTargets(ctx, ctx.RoleId, tasks)

	return &servicepb.Expedition_ExpeditionDispatchResponse{
		Task: task,
	}, nil
}

func (svc *Service) GetDoneReward(ctx *ctx.Context, req *servicepb.Expedition_ExpeditionGetDoneRewardRequest) (*servicepb.Expedition_ExpeditionGetDoneRewardResponse, *errmsg.ErrMsg) {
	e, ok, err := dao.Get(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrExpeditionTaskNotExist()
	}
	task, ok := e.Task[req.TaskId]
	if !ok {
		return nil, errmsg.NewErrExpeditionTaskNotExist()
	}
	if !task.Ongoing || task.DoneTime > timer.StartTime(ctx.StartTime).Unix() {
		return nil, errmsg.NewErrExpeditionTaskNotDone()
	}
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	cfg, ok := rule.GetExpeditionRewardByLevel(ctx, task.ConfigId, role.Level)
	if !ok {
		return nil, errmsg.NewErrExpeditionTaskNotExist()
	}
	if len(cfg.Reward) > 0 {
		if _, err := svc.AddManyItem(ctx, ctx.RoleId, cfg.Reward); err != nil {
			return nil, err
		}
	}
	delete(e.Task, req.TaskId)
	// 生成新的任务
	_, err = svc.genTask(ctx, e, false, false)
	if err != nil {
		return nil, err
	}
	svc.handleFreeRefresh(ctx, e)
	_, _, _, err = svc.handleExecutionRecover(ctx, e)
	if err != nil {
		return nil, err
	}

	dao.Save(ctx, e)

	return &servicepb.Expedition_ExpeditionGetDoneRewardResponse{
		Task:   e.Task,
		Reward: cfg.Reward,
	}, nil
}

func (svc *Service) DoneImmediately(ctx *ctx.Context, req *servicepb.Expedition_ExpeditionDoneImmediatelyRequest) (*servicepb.Expedition_ExpeditionDoneImmediatelyResponse, *errmsg.ErrMsg) {
	e, ok, err := dao.Get(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrExpeditionTaskNotExist()
	}
	task, ok := e.Task[req.TaskId]
	if !ok {
		return nil, errmsg.NewErrExpeditionTaskNotExist()
	}
	if !task.Ongoing {
		return nil, errmsg.NewErrExpeditionTaskNotStart()
	}
	now := timer.StartTime(ctx.StartTime)
	if task.DoneTime <= now.Unix() {
		return nil, errmsg.NewErrExpeditionTaskFinish()
	}
	v := rule.GetExpeditionAccelerateCost(ctx)
	cost := v * values.Integer(math.Ceil(timer.StartTime(task.DoneTime*values.Integer(time.Second)).Sub(now).Minutes()))
	if cost > 0 {
		if err := svc.SubItem(ctx, ctx.RoleId, enum.BoundDiamond, cost); err != nil {
			return nil, err
		}
	}
	task.DoneTime = now.Unix()
	e.Task[req.TaskId] = task

	svc.handleFreeRefresh(ctx, e)
	if _, _, _, err = svc.handleExecutionRecover(ctx, e); err != nil {
		return nil, err
	}

	dao.Save(ctx, e)

	return &servicepb.Expedition_ExpeditionDoneImmediatelyResponse{
		Task: task,
	}, nil
}

func (svc *Service) RefreshTask(ctx *ctx.Context, _ *servicepb.Expedition_ExpeditionRefreshRequest) (*servicepb.Expedition_ExpeditionRefreshResponse, *errmsg.ErrMsg) {
	e, _, err := dao.Get(ctx)
	if err != nil {
		return nil, err
	}
	svc.handleFreeRefresh(ctx, e)
	free := rule.GetExpeditionFreeRefresh(ctx)
	// 先用免费刷新次数
	if e.Refresh.FreeCount < free {
		e.Refresh.FreeCount++
	} else {
		// 没有免费刷新次数的情况下用刷新道具
		itemId := rule.GetExpeditionRefreshItem(ctx)
		count, err := svc.GetItem(ctx, ctx.RoleId, itemId)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			if err := svc.SubItem(ctx, ctx.RoleId, itemId, 1); err != nil {
				return nil, err
			}
		} else {
			// 最后用钻石
			diamond, ok := rule.GetExpeditionRefreshCost(ctx)
			if !ok {
				return nil, errmsg.NewInternalErr("ExpeditionRefreshCost config is nil")
			}
			if err := svc.SubItem(ctx, ctx.RoleId, diamond.ItemId, diamond.Count); err != nil {
				return nil, err
			}
		}
	}
	e.MustCount++
	must := svc.getMustCount(ctx, e) == 0
	if must {
		e.MustCount = 0
	}
	update, err := svc.genTask(ctx, e, true, must)
	if err != nil {
		return nil, err
	}
	_, _, save, err := svc.handleExecutionRecover(ctx, e)
	if err != nil {
		return nil, err
	}
	if update || save {
		dao.Save(ctx, e)
	}

	return &servicepb.Expedition_ExpeditionRefreshResponse{
		MustCount:        svc.getMustCount(ctx, e),
		Task:             e.Task,
		FreeRefreshCount: svc.getFreeRefreshCount(ctx, e),
	}, nil
}

func (svc *Service) handleFreeRefresh(ctx *ctx.Context, e *pbdao.Expedition) bool {
	now := timer.StartTime(ctx.StartTime)
	if e.Refresh.ResetTime == 0 || e.Refresh.ResetTime <= now.Unix() {
		// resetTime := timer.BeginOfDay(now).AddDate(0, 0, 1).Add(time.Second * rule.GetDefaultRefreshTime(ctx))
		e.Refresh = &pbdao.ExpeditionRefresh{
			FreeCount: 0,
			ResetTime: timer.NextDay(svc.GetCurrDayFreshTime(ctx)).Unix(),
		}
		return true
	}
	return false
}

func (svc *Service) genTask(ctx *ctx.Context, e *pbdao.Expedition, refresh, must bool) (bool, *errmsg.ErrMsg) {
	if refresh {
		for id, task := range e.Task {
			if !task.Ongoing {
				delete(e.Task, id)
			}
		}
	}
	count := e.NormalSlot + e.ExtraSlot - values.Integer(len(e.Task))
	// 线上如果策划修改配置，可能会导致可用的slot<当前有的任务数量
	if count <= 0 {
		return false, nil
	}
	all := rule.GetAllExpedition(ctx)
	pool, ok := all[0] // 0是不需要可接条件
	if !ok {
		pool = map[values.Quality][]rule_model.Expedition{}
	}
	taskTypeList := make([]models.TaskType, 0)
	for taskType := range all {
		if taskType == 0 {
			continue
		}
		taskTypeList = append(taskTypeList, taskType)
	}
	taskData, err := svc.GetCounterByTypeList(ctx, taskTypeList)
	if err != nil {
		return false, err
	}
	for taskType, item := range all {
		if taskType == 0 {
			continue
		}
		for q, expedition := range item {
			for _, el := range expedition {
				acceptCondtion := el.AcceptCondtion
				if len(acceptCondtion) != 3 {
					continue
				}
				taskType, condition, need := acceptCondtion[0], acceptCondtion[1], acceptCondtion[2]
				acceptConditionInfo, ok := taskData[models.TaskType(taskType)]
				if !ok {
					continue
				}
				if acceptConditionInfo[condition] >= need {
					if _, ok := pool[q]; !ok {
						pool[q] = make([]rule_model.Expedition, 0)
					}
					pool[q] = append(pool[q], el)
				}
			}
		}
	}

	// pool里是所有可接的任务列表（包含限定条件不满足的）
	for i := 0; i < int(count); i++ {
		task, err := svc.genOneTask(ctx, pool, must)
		if err != nil {
			return false, err
		}
		if must {
			must = false
		}
		e.Task[task.TaskId] = task
	}
	return true, nil
}

func (svc *Service) genOneTask(ctx *ctx.Context, pool map[values.Quality][]rule_model.Expedition, must bool) (*models.ExpeditionTask, *errmsg.ErrMsg) {
	weightMap, ok := rule.GetExpeditionQualityWeight(ctx)
	if !ok {
		return nil, errmsg.NewInternalErr("ExpeditionQualityWeight config not found")
	}
	choices := make([]*wr.Choice[int64, int64], 0)
	if must {
		var maxQuality, weight values.Integer
		for k, v := range weightMap {
			if k > maxQuality {
				maxQuality = k
				weight = v
			}
		}
		choices = append(choices, wr.NewChoice(maxQuality, weight))
	} else {
		for k, v := range weightMap {
			choices = append(choices, wr.NewChoice(k, v))
		}
	}
	chooser, _ := wr.NewChooser(choices...)
	q := chooser.Pick()
	taskList, ok := pool[q]
	if !ok {
		svc.log.Error("get expedition by quality not found", zap.Int64("quality", q))
		return nil, errmsg.NewInternalErr("get expedition by quality not found")
	}
	if len(taskList) <= 0 {
		return nil, errmsg.NewInternalErr("invalid taskList len")
	}
	task := taskList[rand.Intn(len(taskList))]
	return &models.ExpeditionTask{
		TaskId:   xid.New().String(),
		ConfigId: task.Id,
		Ongoing:  false,
		DoneTime: 0,
	}, nil
}

// 获取任务栏数量（不包含额外的，仅expedition_quantity配置表里解锁的栏位）
func (svc *Service) getSlotCount(ctx *ctx.Context) (values.Integer, *errmsg.ErrMsg) {
	data := rule.GetAllExpeditionQuantity(ctx)
	noLimit, ok := data[0]
	var count values.Integer
	if ok {
		sort.Slice(noLimit, func(i, j int) bool {
			return noLimit[i].Id > noLimit[j].Id
		})
		// 不需要解锁条件的数量
		count = noLimit[0].Id
	}

	temp := make([]rule_model.ExpeditionQuantity, 0)
	taskTypeList := make([]models.TaskType, 0)
	for taskType := range data {
		if taskType == 0 {
			continue
		}
		taskTypeList = append(taskTypeList, taskType)
	}
	taskData, err := svc.GetCounterByTypeList(ctx, taskTypeList)
	if err != nil {
		return 0, err
	}
	for taskType, item := range data {
		if taskType == 0 {
			continue
		}
		for _, eq := range item {
			unlockConditionCount, ok := taskData[models.TaskType(eq.UnlockCondtion[0])]
			if !ok {
				continue
			}
			if unlockConditionCount[0] >= eq.UnlockCondtion[2] {
				temp = append(temp, eq)
			}
		}
	}
	if len(temp) > 0 {
		sort.Slice(temp, func(i, j int) bool {
			return temp[i].Id > temp[j].Id
		})
		curCount := temp[0].Id
		if curCount > count {
			count = curCount
		}
	}
	return count, nil
}

func (svc *Service) getFreeRefreshCount(ctx *ctx.Context, e *pbdao.Expedition) values.Integer {
	v := rule.GetExpeditionFreeRefresh(ctx)
	count := v - e.Refresh.FreeCount
	if count < 0 {
		count = 0
	}
	return count
}

func (svc *Service) getMustCount(ctx *ctx.Context, e *pbdao.Expedition) values.Integer {
	v := rule.GetExpeditionRefreshFloors(ctx)
	count := v - e.MustCount - 1
	if count < 0 {
		count = 0
	}
	return count
}

// 处理行动力的恢复逻辑（返回：单个恢复间隔恢复的数量、下次恢复时间、是否需要更新db）
func (svc *Service) handleExecutionRecover(ctx *ctx.Context, e *pbdao.Expedition) (values.Integer, values.Integer, bool, *errmsg.ErrMsg) {
	interval, num := rule.GetExpeditionCostRecovery(ctx)
	now := timer.StartTime(ctx.StartTime)
	for _, v := range e.Execution.ExtraRestoreCount {
		num += v
	}
	lastRecover := timer.StartTime(e.Execution.LastRecoverTime * values.Integer(time.Second))
	if now.Sub(lastRecover).Minutes() < interval {
		return num, lastRecover.Add(time.Duration(interval) * time.Minute).Unix(), false, nil
	}
	haveCount, err := svc.GetItem(ctx, ctx.RoleId, rule.GetExpeditionTaskCostItemId(ctx))
	if err != nil {
		return 0, 0, false, err
	}
	max := svc.getMaxExecution(ctx, e.Execution)
	need := max - haveCount
	if need <= 0 {
		e.Execution.LastRecoverTime = now.Unix()
		return num, now.Unix(), true, nil
	}
	var count values.Integer
	for now.Sub(lastRecover).Minutes() >= interval {
		count++
		lastRecover = lastRecover.Add(time.Duration(interval) * time.Minute)
	}
	total := num * count
	if total <= 0 {
		return num, lastRecover.Add(time.Duration(interval) * time.Minute).Unix(), false, nil
	}
	var add values.Integer
	if total <= need {
		add = total
	} else {
		add = need
	}
	if err := svc.AddItem(ctx, ctx.RoleId, rule.GetExpeditionTaskCostItemId(ctx), add); err != nil {
		return 0, 0, false, err
	}
	e.Execution.LastRecoverTime = lastRecover.Unix()
	return num, lastRecover.Add(time.Duration(interval) * time.Minute).Unix(), true, nil
}

func (svc *Service) nilCheck(e *pbdao.Expedition) {
	if e.Task == nil {
		e.Task = map[string]*models.ExpeditionTask{}
	}
	if e.Execution == nil {
		e.Execution = &pbdao.Execution{}
	}
	if e.Execution.ExtraRestoreCount == nil {
		e.Execution.ExtraRestoreCount = map[int64]int64{}
	}
	if e.Execution.LimitBonus == nil {
		e.Execution.LimitBonus = map[int64]int64{}
	}
	if e.Refresh == nil {
		e.Refresh = &pbdao.ExpeditionRefresh{}
	}
}

func (svc *Service) getMaxExecution(ctx *ctx.Context, e *pbdao.Execution) values.Integer {
	max := rule.GetExpeditionCostLimit(ctx)
	if e != nil && e.LimitBonus != nil {
		for _, val := range e.LimitBonus {
			max += val
		}
	}
	return max
}

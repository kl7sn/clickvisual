package service

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/ego-component/egorm"
	"github.com/gotomicro/ego/core/elog"

	"github.com/clickvisual/clickvisual/api/internal/invoker"
	db2 "github.com/clickvisual/clickvisual/api/internal/pkg/model/db"
	"github.com/clickvisual/clickvisual/api/internal/pkg/model/view"
	"github.com/clickvisual/clickvisual/api/internal/service/alarm/pusher"
	"github.com/clickvisual/clickvisual/api/internal/service/inquiry"
)

func (i *alert) HandlerAlertManager(alarmUUID string, filterIdStr string, notification db2.Notification) (err error) {
	// 获取告警信息
	tx := invoker.Db.Begin()
	conds := egorm.Conds{}
	conds["uuid"] = alarmUUID
	alarm, err := db2.AlarmInfoX(tx, conds)
	if err != nil {
		tx.Rollback()
		return err
	}
	if alarm.Status == db2.AlarmStatusClose {
		tx.Commit()
		return
	}
	notifStatus := notification.GetStatus() // 当前需要推送的状态
	if alarm.IsDisableResolve == 1 && notifStatus == db2.AlarmStatusNormal {
		tx.Commit()
		return
	}
	// create history
	filterId, _ := strconv.Atoi(filterIdStr)
	alarmHistory := db2.AlarmHistory{AlarmId: alarm.ID, FilterId: filterId, FilterStatus: notifStatus, IsPushed: db2.PushedStatusRepeat}
	if err = db2.AlarmHistoryCreate(tx, &alarmHistory); err != nil {
		tx.Rollback()
		return err
	}
	currentFiltersStatus := alarm.GetStatus(tx)
	// update filter
	af := db2.AlarmFilter{}
	af.ID = filterId
	af.Status = notifStatus
	if err = af.UpdateStatus(tx); err != nil {
		tx.Rollback()
		return
	}
	if err = alarm.UpdateStatus(tx, alarm.GetStatus(tx)); err != nil {
		elog.Error("PushAlertManagerError", elog.FieldErr(err), elog.Int("filterId", filterId), elog.String("alarmUUID", alarmUUID))
		tx.Rollback()
		return err
	}
	if currentFiltersStatus == notifStatus && time.Now().Unix()-alarm.Utime < 300 {
		// 此时有正在进行中的告警
		elog.Info("PushAlertManagerRepeat", elog.Int("notifStatus", notifStatus), elog.Int("filterId", filterId), elog.String("alarmUUID", alarmUUID))
		tx.Commit()
		return nil
	}
	// 完成告警状态更新
	tx.Commit()
	// get alarm filter info
	filter, err := i.compatibleFilter(alarm.ID, filterId)
	if err != nil {
		elog.Error("PushAlertManagerError", elog.FieldErr(err), elog.Int("filterId", filterId), elog.String("alarmUUID", alarmUUID))
		return
	}
	// get table info
	tableInfo, err := db2.TableInfo(invoker.Db, filter.Tid)
	if err != nil {
		elog.Error("PushAlertManagerError", elog.FieldErr(err), elog.Int("filterId", filterId), elog.String("alarmUUID", alarmUUID))
		return
	}
	if tableInfo.TimeField == "" {
		tableInfo.TimeField = db2.TimeFieldSecond
	}
	// get op
	op, err := InstanceManager.Load(tableInfo.Database.Iid)
	if err != nil {
		elog.Error("PushAlertManagerError", elog.FieldErr(err), elog.Int("filterId", filterId), elog.String("alarmUUID", alarmUUID))
		return
	}
	// get partial log
	partialLog := i.getPartialLog(op, &tableInfo, &alarm, filter)

	pushMsg, err := pusher.BuildAlarmMsg(notification, &tableInfo, &alarm, filter, partialLog)
	if err != nil {
		elog.Error("PushAlertManagerError", elog.FieldErr(err), elog.Int("filterId", filterId), elog.String("alarmUUID", alarmUUID))
		return
	}

	pushMsgWithAt, err := pusher.BuildAlarmMsgWithAt(notification, &tableInfo, &alarm, filter, partialLog)
	if err != nil {
		elog.Error("PushAlertManagerError", elog.FieldErr(err), elog.Int("filterId", filterId), elog.String("alarmUUID", alarmUUID))
		return
	}

	if err = pusher.Execute(alarm.ChannelIds, pushMsg, pushMsgWithAt); err != nil {
		elog.Error("PushAlertManagerError", elog.FieldErr(err), elog.Int("filterId", filterId), elog.String("alarmUUID", alarmUUID))
		_ = db2.AlarmHistoryUpdate(invoker.Db, alarmHistory.ID, map[string]interface{}{"is_pushed": db2.PushedStatusFail})
		return
	}
	if err = db2.AlarmHistoryUpdate(invoker.Db, alarmHistory.ID, map[string]interface{}{"is_pushed": db2.PushedStatusSuccess}); err != nil {
		return err
	}
	return nil
}

func (i *alert) compatibleFilter(alarmId int, filterId int) (res *db2.AlarmFilter, err error) {
	if filterId == 0 {
		condsFilter := egorm.Conds{}
		condsFilter["alarm_id"] = alarmId
		filters, errAlarmFilterList := db2.AlarmFilterList(invoker.Db, condsFilter)
		if errAlarmFilterList != nil {
			return nil, errAlarmFilterList
		}
		if len(filters) == 0 {
			return nil, errors.New("empty alarm filter")
		}
		res = filters[0]
	} else {
		filter, errAlarmFilterInfo := db2.AlarmFilterInfo(invoker.Db, filterId)
		if errAlarmFilterInfo != nil {
			return nil, errAlarmFilterInfo
		}
		res = &filter
	}
	return
}

func (i *alert) getPartialLog(op inquiry.Operator, table *db2.BaseTable, alarm *db2.Alarm, filter *db2.AlarmFilter) (partialLog string) {
	param := view.ReqQuery{
		Tid:           table.ID,
		Database:      table.Database.Name,
		Table:         table.Name,
		Query:         filter.When,
		AlarmMode:     filter.Mode,
		TimeField:     table.TimeField,
		TimeFieldType: table.TimeFieldType,
		ST:            time.Now().Add(-alarm.GetInterval() - time.Minute).Unix(),
		ET:            time.Now().Add(time.Minute).Unix(),
		Page:          1,
		PageSize:      1,
	}
	param, _ = op.Prepare(param, false)
	resp, _ := op.GetLogs(param, table.ID)
	if table.V3TableType == db2.V3TableTypeJaegerJSON {
		resp.IsTrace = 1
	}
	if len(resp.Logs) > 0 {
		l, _ := json.Marshal(resp.Logs[0])
		partialLog = string(l)
	}
	return partialLog
}

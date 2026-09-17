package snowflake

import (
	"fmt"
	"time"

	"github.com/sony/sonyflake"
)

var (
	sonyFlake     *sonyflake.Sonyflake
	sonyMachineID uint16
	// startTime 雪花算法的起始时间, 决定了 ID 的位数以及可用年限
	startTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
)

func getMachineID() (uint16, error) {
	return sonyMachineID, nil
}

// Init 需传入当前的机器ID
func Init(machineID uint16) (err error) {
	sonyMachineID = machineID
	settings := sonyflake.Settings{
		StartTime: startTime,
		MachineID: getMachineID,
	}
	sonyFlake = sonyflake.NewSonyflake(settings)
	if sonyFlake == nil {
		return fmt.Errorf("init sonyflake failed, machine_id=%d", machineID)
	}
	return nil
}

// GetID 返回生成的id值
func GetID() (id uint64, err error) {
	if sonyFlake == nil {
		err = fmt.Errorf("sonyflake not inited")
		return
	}

	id, err = sonyFlake.NextID()
	return
}

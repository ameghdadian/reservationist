package appointment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ameghdadian/service/business/core/user"
	"github.com/ameghdadian/service/foundation/logger"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypeSendSMS = "sms:send"
)

type Task struct {
	client    *asynq.Client
	inspector *asynq.Inspector
}

func NewTask(client *asynq.Client, inspector *asynq.Inspector) *Task {
	return &Task{
		client:    client,
		inspector: inspector,
	}
}

type sendSMSPayload struct {
	UserID uuid.UUID
}

func (t *Task) NewSendSMSTask(userID uuid.UUID, fireAt time.Time, taskID string) (*asynq.TaskInfo, error) {
	fireAt = fireAt.UTC()
	data := sendSMSPayload{UserID: userID}
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("creating a new send sms task: %w", err)
	}

	task := asynq.NewTask(
		TypeSendSMS,
		payload,
		asynq.TaskID(taskID),
		asynq.ProcessAt(fireAt),
		asynq.Timeout(time.Minute*1),
	)

	info, err := t.client.Enqueue(task)
	if err != nil {
		return nil, fmt.Errorf("enqueue task[%s]: %w", TypeSendSMS, err)
	}

	return info, err
}

func (t *Task) cancelSendSMSTask(taskID string) error {
	if err := t.inspector.DeleteTask("default", taskID); err != nil {
		return fmt.Errorf("delete scheduled sms task: %w", err)
	}

	return nil
}

func (t *Task) updateSendSMSTask(userID uuid.UUID, newFireAt time.Time, taskID string) (*asynq.TaskInfo, error) {
	if err := t.cancelSendSMSTask(taskID); err != nil {
		return nil, fmt.Errorf("delete scheduled sms task: %w", err)
	}

	ti, err := t.NewSendSMSTask(userID, newFireAt, taskID)
	if err != nil {
		return nil, err
	}

	return ti, nil
}

// ----------------------------------------------------------------------------------------------------------

type TaskHandlers struct {
	log     *logger.Logger
	usrCore *user.Core
}

func NewTaskHandlers(log *logger.Logger, usrCore *user.Core) *TaskHandlers {
	return &TaskHandlers{
		log:     log,
		usrCore: usrCore,
	}
}

func (t *TaskHandlers) HandleSendSMS(ctx context.Context, tsk *asynq.Task) error {
	var s sendSMSPayload
	if err := json.Unmarshal(tsk.Payload(), &s); err != nil {
		return err
	}

	// Query user domain using core objects in TaskHandlers

	// Simulate sending an early notification(reminder) to user
	t.log.Info(ctx, "sending reminder to userID[%s] at[%s]", s.UserID)

	return nil
}

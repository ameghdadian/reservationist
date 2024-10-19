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
	client *asynq.Client
}

func NewTask(client *asynq.Client) *Task {
	return &Task{
		client: client,
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

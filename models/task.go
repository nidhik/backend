package models

import (
	"fmt"

	"github.com/nidhik/backend/db"
	"gopkg.in/mgo.v2/bson"
)

const (
	HandleEmail   = "sendEmail"
	TaskTypeEmail = "TRNS_EMAIL_GO"

	HandleExpirePro   = "expireProSubscription"
	TaskTypeExpirePro = "EXPIRE_SUBSCRIPTION_GO"

	HandleSubscriptionCreated = "handleSubscriptionCreated"
	HandleSuccessfulPayment   = "handleSuccessfulPayment"
	HandleFailedPayment       = "handleFailedPayment"
	HandleSubscriptionDeleted = "handleSubscriptionDeleted"
	TaskTypeHandleStripeEvent = "HANDLE_STRIPE_EVENT_GO"

	HandleApplyReferralDiscount   = "handleApplyReferralDiscount"
	TaskTypeApplyReferralDiscount = "APPLY_REFERRAL_GO"
)

const (
	CollectionTask = "Task"
)

type Task struct {
	Message      string        `json:"taskMessage" bson:"taskMessage"`
	Status       string        `json:"taskStatus" bson:"taskStatus"`
	Action       string        `json:"taskAction" bson:"taskAction"`
	Type         string        `json:"taskType" bson:"taskType"`
	Parameters   []interface{} `json:"taskParameters" bson:"taskParameters"`
	Claimed      int           `json:"taskClaimed" bson:"taskClaimed"`
	User         *User         `json:"user" bson:"-"`
	UserPtr      string        `json:"-" bson:"_p_user"`
	db.BaseModel `bson:",inline"`
}

func NewTask(id string) *Task {
	return &Task{
		BaseModel: db.BaseModel{
			Id:             id,
			CollectionName: CollectionTask},
	}
}

func NewEmptyTask() *Task {
	return &Task{
		BaseModel: db.BaseModel{
			CollectionName: CollectionTask},
	}
}

func NewTaskForUser(user *User, taskType string, taskAction string, taskParameters ...interface{}) *Task {
	userPtr := user.Collection() + "$" + user.ObjectId()

	task := NewEmptyTask()

	task.User = user
	task.Set("UserPtr", userPtr)
	task.Set("Type", taskType)
	task.Set("Action", taskAction)
	task.Set("Parameters", taskParameters)
	task.Set("Status", "NEW")
	task.Set("Claimed", 0)

	return task

}

func (task *Task) Fetch(ds db.DataStore) error {
	return task.BaseModel.Fetch(task, ds)
}

func (task *Task) Save(ds db.DataStore) error {
	return task.BaseModel.Save(task, ds)
}

func (task *Task) Delete(ds db.DataStore) error {
	return task.BaseModel.Delete(task, ds)
}

func (task *Task) Set(fieldName string, value interface{}) {
	task.BaseModel.Set(task, fieldName, value)
}

func (task *Task) Unset(fieldName string) {
	task.BaseModel.Unset(task, fieldName)
}

func (task *Task) Get(fieldName string) interface{} {
	return task.BaseModel.Get(task, fieldName)
}

func (task *Task) Increment(fieldName string, amount int) {
	task.BaseModel.Increment(task, fieldName, amount)
}

func (model *Task) CustomUnmarshall() {

	if ptr := UnmarshallPointer(model.UserPtr); ptr != nil {
		if model.User == nil || model.User.ObjectId() != ptr.ObjectId {
			m, _ := ptr.model()
			model.User = m.(*User)
		}
	}

	model.CollectionName = CollectionTask
}

// Methods specific to Task

func (task *Task) FetchParameters(ds db.DataStore) ([]interface{}, *User, error) {
	params := task.Parameters
	var loadedParams []interface{}

	for _, param := range params {

		switch param.(type) {
		case bson.M:
			ptr := param.(bson.M)
			if ptr["__type"] == "Pointer" {
				newParam := NewPointer(ptr["className"].(string), ptr["objectId"].(string))
				if m, err := newParam.Fetch(ds); err != nil {
					fmt.Printf("Error loading params %s Param: %s \n", err, newParam)
					return nil, nil, err
				} else {
					loadedParams = append(loadedParams, m)
				}

			} else {
				loadedParams = append(loadedParams, param)
			}

		default:
			loadedParams = append(loadedParams, param)

		}
	}

	if err := task.User.Fetch(ds); err != nil {
		return nil, nil, err
	}

	return loadedParams, task.User, nil
}

func FindEachClaimedTask(ds db.DataStore, f func(task *Task)) error {
	fmt.Printf("Finding completed tasks.\n")

	return ds.FindEach(CollectionTask, bson.M{"taskClaimed": 1}, func(model db.Model) {
		var t = model.(*Task)
		ptr := NewEmptyTask()
		*ptr = *t
		f(ptr)

	}, &Task{})

}

func FindEachNewTask(ds db.DataStore, f func(task *Task)) error {
	fmt.Printf("Finding new tasks.\n")

	return ds.FindEach(CollectionTask, bson.M{"taskClaimed": 0, "taskType": bson.M{"$in": []string{TaskTypeEmail, TaskTypeExpirePro, TaskTypeHandleStripeEvent, TaskTypeApplyReferralDiscount}}}, func(model db.Model) {
		var t = model.(*Task)
		ptr := NewEmptyTask()
		*ptr = *t
		f(ptr)

	}, &Task{})

}

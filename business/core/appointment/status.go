package appointment

import "fmt"

var (
	Scheduled = Status{"Scheduled"}
	Cancelled = Status{"Cancelled"}
)

var statuses = map[string]Status{
	Cancelled.status: Cancelled,
	Scheduled.status: Scheduled,
}

type Status struct {
	status string
}

func ParseStatus(value string) (Status, error) {
	status, exists := statuses[value]
	if !exists {
		return Status{}, fmt.Errorf("invalid status: %q", value)
	}

	return status, nil
}

func (as Status) Status() string {
	return as.status
}

func (as *Status) UnmarshalText(data []byte) error {
	status, err := ParseStatus(string(data))
	if err != nil {
		return err
	}

	as.status = status.status
	return nil
}

func (as Status) MarshalText() ([]byte, error) {
	return []byte(as.status), nil
}

func (as Status) Equal(as2 Status) bool {
	return as.status == as2.status
}

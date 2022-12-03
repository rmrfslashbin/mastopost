package events

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

// AWSRegionRequiredError is returned when AWS Region is not set
type AWSRegionRequiredError struct {
	Err error
}

// Error returns the error message
func (e *AWSRegionRequiredError) Error() string {
	if e.Err == nil {
		return "AWS Region is required. Use WithRegion() to set it."
	}
	return e.Err.Error()
}

// AWSConfigError is an error returned when there is an error with the AWS Config
type AWSConfigError struct {
	Err error
}

// Error returns the error message
func (e *AWSConfigError) Error() string {
	if e.Err == nil {
		return "AWS Config error"
	} else {
		return fmt.Sprintf("AWS Config error: %s", e.Err.Error())
	}
}

// DeleteRuleError is an error returned when there is an error deleting a rule
type DeleteRuleError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *DeleteRuleError) Error() string {
	if e.Msg == "" {
		e.Msg = "error deleting AWS rule"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// DescribeRuleError is an error returned when there is an error with the DescribeRule API call
type DescribeRuleError struct {
	Err error
}

// Error returns the error message
func (e *DescribeRuleError) Error() string {
	if e.Err == nil {
		return "EventBridge DescribeRule API call error"
	} else {
		return fmt.Sprintf("EventBridge DescribeRule API call error: %s", e.Err.Error())
	}
}

// DisableRuleError is an error returned when there is an error disabling a rule
type DisableRuleError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *DisableRuleError) Error() string {
	if e.Msg == "" {
		e.Msg = "error disabling AWS rule"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// EnableRuleError is an error returned when there is an error enabling a rule
type EnableRuleError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *EnableRuleError) Error() string {
	if e.Msg == "" {
		e.Msg = "error enabling AWS rule"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// PutRuleError is an error returned when there is an error with the PutRule call
type PutRuleError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *PutRuleError) Error() string {
	if e.Msg == "" {
		e.Msg = "error putting event bridge rule"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// RemovePermissionError is an error returned when there is an error with the RemovePermission call
type RemovePermissionError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *RemovePermissionError) Error() string {
	if e.Msg == "" {
		e.Msg = "error removing lambda permission"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// AddPermissionError is an error returned when there is an error with the AddPermission call
type AddPermissionError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *AddPermissionError) Error() string {
	if e.Msg == "" {
		e.Msg = "error adding IAM permission to Lambda function"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// PutTargetsError is an error returned when there is an error with the PutTargets call
type PutTargetsError struct {
	Err              error
	Msg              string
	FailedEntryCount *int32
	FailedEntries    *[]types.PutTargetsResultEntry
}

// Error returns the error message
func (e *PutTargetsError) Error() string {
	if e.Msg == "" {
		e.Msg = "error adding event bridge rule target"
	}
	if e.FailedEntryCount != nil {
		e.Msg += fmt.Sprintf(": Failed Entry Count: %d", *e.FailedEntryCount)
	}
	if e.FailedEntries != nil {
		e.Msg += fmt.Sprintf(": Failed Entries: %v", e.FailedEntries)
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

type RemoveTargetsError struct {
	Err              error
	Msg              string
	FailedEntries    *[]types.RemoveTargetsResultEntry
	FailedEntryCount *int32
}

func (e *RemoveTargetsError) Error() string {
	if e.Msg == "" {
		e.Msg = "error removing event bridge rule target"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

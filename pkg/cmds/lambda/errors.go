package lambda

import "fmt"

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

type EventBridgeDeleteError struct {
	Msg string
	Err error
}

// Error returns the error message
func (e *EventBridgeDeleteError) Error() string {
	if e.Msg == "" {
		e.Msg = "error deleting AWS EventBridge rule"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// FeedLoadError is returned when a feed cannot be loaded
type FeedLoadError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *FeedLoadError) Error() string {
	if e.Msg == "" {
		e.Msg = "error loading feed"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// FeedNameMissingError is returned when the feed name is missing
type FeedNameMissingError struct {
	Err error
}

// Error returns the error message
func (e *FeedNameMissingError) Error() string {
	if e.Err == nil {
		return "Feed name is required"
	}
	return e.Err.Error()
}

// FeedNotInConfig is returned when a feed is not in the config
type FeedNotInConfig struct {
	Err      error
	Msg      string
	feedname string
}

// Error returns the error message
func (e *FeedNotInConfig) Error() string {
	if e.Msg == "" {
		e.Msg = "feed not in config"
	}
	if e.feedname != "" {
		e.Msg += ": " + e.feedname
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// NoConfigFile is returned when a filename is required but not provided
type NoConfigFile struct {
	Err error
}

// Error returns the error message
func (e *NoConfigFile) Error() string {
	if e.Err == nil {
		return "no config file provided. use WithConfigFile() to set the config file"
	}
	return e.Err.Error()
}

// NoConfirm is returned when a confirmation is required but not provided
type NoConfirm struct {
	Err error
}

// Error returns the error message
func (e *NoConfirm) Error() string {
	if e.Err == nil {
		return "user rejected confirmation"
	}
	return e.Err.Error()
}

// NoFeedName is returned when a feed name is required but not provided
type NoFeedName struct {
	Err error
}

// Error returns the error message
func (e *NoFeedName) Error() string {
	if e.Err == nil {
		return "no feed name provided. use WithFeedName() to set the feed name"
	}
	return e.Err.Error()
}

// NoLambdaFunction is returned when the lambda function name is not set
type NoLambdaFunction struct {
	Err error
}

// Error returns the error message
func (e *NoLambdaFunction) Error() string {
	if e.Err == nil {
		return "no lambda function name provided. use WithLambdaFunctionName()"
	}
	return e.Err.Error()
}

// NoLambdaFunctionARN is returned when the lambda function ARN is not set
type NoLambdaFunctionARN struct {
	Err                error
	Msg                string
	LambdaFunctionName *string
}

// Error returns the error message
func (e *NoLambdaFunctionARN) Error() string {
	if e.Msg == "" {
		e.Msg = "can't find lambda function in config"
	}
	if e.LambdaFunctionName != nil {
		e.Msg += ": " + *e.LambdaFunctionName
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// NotImplementedError is returned when a function is not implemented
type NotImplementedError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *NotImplementedError) Error() string {
	if e.Msg == "" {
		e.Msg = "not implemented"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// ParametersDeleteError is returned when there is an error deleting parameters
type ParametersDeleteError struct {
	Err               error
	Msg               string
	InvalidParameters []string
}

// Error returns the error message
func (e *ParametersDeleteError) Error() string {
	if e.Msg == "" {
		e.Msg = "error deleting parameters"
	}
	if len(e.InvalidParameters) > 0 {
		e.Msg += ": " + fmt.Sprintf("%v", e.InvalidParameters)
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

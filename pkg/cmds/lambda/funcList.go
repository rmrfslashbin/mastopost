package lambda

import (
	"fmt"

	"github.com/rmrfslashbin/mastopost/pkg/ssmparams"
)

// List lists the lambda event configurations
func (l *LambdaConfig) List() error {
	params, err := ssmparams.New(
		ssmparams.WithLogger(l.log),
		ssmparams.WithProfile(*l.awsprofile),
		ssmparams.WithRegion(*l.awsregion),
	)
	if err != nil {
		return err
	}
	path := "/mastopost/"
	if l.feedName != nil {
		path = fmt.Sprintf("%s%s/", path, *l.feedName)
	}

	fmt.Printf("Listing parameters for path %s\n", path)
	var nextToken *string
	for {
		opt, err := params.ListAllParams(path, nextToken)
		if err != nil {
			return err
		}

		for _, p := range opt.Parameters {
			fmt.Printf("Name:    %s\n", *p.Name)
			fmt.Printf("Value:   %s\n", *p.Value)
			fmt.Printf("mtime:   %s\n", *p.LastModifiedDate)
			fmt.Printf("Version: %d\n", p.Version)
			fmt.Printf("ARN:     %s\n", *p.ARN)
			fmt.Println()
		}

		nextToken = opt.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

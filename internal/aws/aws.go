// Copyright © 2025 Steve Taranto staranto@gmail.com
// SPDX-License-Identifier: MIT

package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func CreateAWSSession(region string) (*session.Session, error) {
	awsCfg := aws.Config{
		Region: aws.String(region),
	}

	return session.NewSessionWithOptions(session.Options{
		Config:            awsCfg,
		SharedConfigState: session.SharedConfigEnable,
	})
}

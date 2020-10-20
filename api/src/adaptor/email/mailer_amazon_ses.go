package email

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

type MailerAmazonSES struct {
	encoding string
	config   *aws.Config
	configurationSet string
}

// NewMailerAmazonSES is MailerAmazonSES's constructor
func NewMailerAmazonSES(region string, configurationSet string) *MailerAmazonSES {
	m := &MailerAmazonSES{
		encoding: "UTF-8",
		configurationSet: configurationSet,
	}
	// defaults config get AWS region from AWS_REGION
	m.config = defaults.Config().WithRegion(region)
	// custom config:
	// credentials based on env: AWS_ACCESS_KEY / AWS_SECRET_KEY
	m.config.Credentials = credentials.NewEnvCredentials()
	return m
}

// Send uses amazon ses sdk to send an email
func (m *MailerAmazonSES) Send(ctx context.Context, email *EmailNotification) error {
	recipient := email.To

	// Create a new AWS session in the region.
	awsSession, err := session.NewSession(m.config)
	if err != nil {
		return fmt.Errorf("could not create aws session (%v)", err)
	}

	sesService := ses.New(awsSession)

	// Assemble the email.
	input := &ses.SendEmailInput{
		ConfigurationSetName: &m.configurationSet,
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{&recipient},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(m.encoding),
					Data:    aws.String(email.HTMLBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(m.encoding),
				Data:    aws.String(email.Subject),
			},
		},
		Source: aws.String(email.From),
	}

	// Attempt to send the email.
	_, err = sesService.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok {
			switch awsErr.Code() {
			case ses.ErrCodeMessageRejected:
				return fmt.Errorf("aws %s (%v)", ses.ErrCodeMessageRejected, awsErr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				return fmt.Errorf("aws %s (%v)", ses.ErrCodeMailFromDomainNotVerifiedException, awsErr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				return fmt.Errorf("aws %s (%v)", ses.ErrCodeConfigurationSetDoesNotExistException, awsErr.Error())
			default:
				return fmt.Errorf("aws Error (%v)", awsErr.Error())
			}
		}
		return fmt.Errorf("error (%v)", err.Error())
	}

	return nil
}

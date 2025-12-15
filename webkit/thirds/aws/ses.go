package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

var CHARSET = "UTF-8"

func (c *Client) SendEmail(ctx context.Context, emailSender, fromName, html, text, subject string, toAddresses, bccAddresses []string) error {
	fromEmailAddress := fmt.Sprintf("%q <%s>", fromName, emailSender)
	destination := &types.Destination{ToAddresses: toAddresses}
	if bccAddresses != nil && len(bccAddresses) > 0 {
		destination.BccAddresses = bccAddresses
	}

	input := sesv2.SendEmailInput{
		Content: &types.EmailContent{
			Simple: &types.Message{
				Body: &types.Body{
					Html: &types.Content{
						Data:    &html,
						Charset: &CHARSET,
					},
					Text: &types.Content{
						Data:    &text,
						Charset: &CHARSET,
					},
				},
				Subject: &types.Content{
					Data:    &subject,
					Charset: &CHARSET,
				},
			},
		},
		Destination:      destination,
		FromEmailAddress: &fromEmailAddress,
	}
	_, err := c.sesClient.SendEmail(ctx, &input)
	return err
}

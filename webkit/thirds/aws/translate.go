package aws

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/translate"
	"github.com/aws/aws-sdk-go-v2/service/translate/types"
)

func (c *Client) TranslateText(ctx context.Context, sourceLanguageCode, targetLanguageCode, text string) (string, error) {
	if sourceLanguageCode == "" || targetLanguageCode == "" || text == "" {
		return "", nil
	}

	params := &translate.TranslateTextInput{
		SourceLanguageCode: &sourceLanguageCode,
		TargetLanguageCode: &targetLanguageCode,
		Text:               &text,
	}
	output, err := c.translateClient.TranslateText(ctx, params)
	if err != nil {
		return "", err
	}
	if output.TranslatedText == nil {
		return "", errors.New("translate result is empty")
	}

	return *output.TranslatedText, nil
}

func (c *Client) TranslateDocument(ctx context.Context, sourceLanguageCode, targetLanguageCode string, content []byte) ([]byte, error) {
	if sourceLanguageCode == "" || targetLanguageCode == "" || len(content) == 0 {
		return nil, nil
	}

	params := &translate.TranslateDocumentInput{
		SourceLanguageCode: &sourceLanguageCode,
		TargetLanguageCode: &targetLanguageCode,
		Document: &types.Document{
			Content:     content,
			ContentType: aws.String("text/html"),
		},
		Settings: &types.TranslationSettings{
			Formality: types.FormalityFormal,
		},
	}
	output, err := c.translateClient.TranslateDocument(ctx, params)
	if err != nil {
		return nil, err
	}
	if output.TranslatedDocument == nil {
		return nil, errors.New("translate result is empty")
	}

	return output.TranslatedDocument.Content, nil
}

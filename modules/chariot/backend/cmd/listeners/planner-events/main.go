package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/praetorian-inc/chariot/backend/pkg/cloud"
	"github.com/praetorian-inc/tabularium/pkg/model/model"
	"log/slog"
)

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, record := range sqsEvent.Records {
		if err := processMessage(ctx, record.Body); err != nil {
			slog.Error("failed to process message", "error", err, "messageId", record.MessageId)
		}
	}
	return nil
}

func processMessage(ctx context.Context, body string) error {
	var event model.PlannerEvent
	if err := json.Unmarshal([]byte(body), &event); err != nil {
		return fmt.Errorf("failed to unmarshal planner event: %w", err)
	}
	
	return processJobCompletion(ctx, event)
}

func processJobCompletion(ctx context.Context, event model.PlannerEvent) error {
	var builder strings.Builder
	
	if event.Status == "JF" {
		builder.WriteString("🔴 **Security Scan Failed**\n\n")
		builder.WriteString(fmt.Sprintf("**Capability**: %s\n", event.Source))
		builder.WriteString(fmt.Sprintf("**Target**: %s\n", event.Target))
		builder.WriteString(fmt.Sprintf("**Error**: %s\n", event.Comment))
		builder.WriteString("**Recommendation**: Check target accessibility and try again.")
	} else {
		builder.WriteString("🎯 **Security Scan Complete**\n\n")
		builder.WriteString(fmt.Sprintf("**Capability**: %s\n", event.Source))
		builder.WriteString(fmt.Sprintf("**Target**: %s\n", event.Target))
		builder.WriteString(fmt.Sprintf("**Total Findings**: %d\n", event.TotalCount))
		
		if len(event.ResultKeys) > 0 {
			builder.WriteString("\n**Key Findings**:\n")
			for i, key := range event.ResultKeys {
				if i >= 5 {
					builder.WriteString(fmt.Sprintf("... and %d more findings\n", len(event.ResultKeys)-5))
					break
				}
				builder.WriteString(fmt.Sprintf("- %s\n", key))
			}
		}
		
		if event.TotalCount > len(event.ResultKeys) {
			builder.WriteString(fmt.Sprintf("\n*Total: %d findings (%d additional not shown)*", event.TotalCount, event.TotalCount-len(event.ResultKeys)))
		}
	}
	
	summary := builder.String()
	
	aws := cloud.NewAWS(event.Username)
	aiMsg := model.NewMessage(event.ConversationID, model.RoleChariot, summary, "system")
	
	if err := aws.Table.Insert(&aiMsg); err != nil {
		return fmt.Errorf("failed to store job completion summary: %w", err)
	}
	
	slog.Info("processed job completion", 
		"conversationId", event.ConversationID, 
		"jobKey", event.JobKey,
		"status", event.Status,
		"totalResults", event.TotalCount)
	return nil
}

func main() {
	lambda.Start(handler)
}
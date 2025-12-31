package rag

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type Indexer struct {
	service *Service
}

func NewIndexer(service *Service) *Indexer {
	if service == nil {
		return nil
	}
	return &Indexer{service: service}
}

func (i *Indexer) IndexTask(ctx context.Context, orgID, taskID uuid.UUID, title, description string) {
	if i == nil || i.service == nil {
		return
	}

	content := fmt.Sprintf("Task: %s\n\n%s", title, description)
	if content == "" {
		return
	}

	err := i.service.IndexDocument(ctx, IndexRequest{
		OrgID:      orgID,
		SourceType: "task",
		SourceID:   taskID,
		Content:    content,
	})
	if err != nil {
		log.Printf("Warning: failed to index task %s: %v", taskID, err)
	}
}

func (i *Indexer) IndexIssue(ctx context.Context, orgID, issueID uuid.UUID, title, description string) {
	if i == nil || i.service == nil {
		return
	}

	content := fmt.Sprintf("Issue: %s\n\n%s", title, description)
	if content == "" {
		return
	}

	err := i.service.IndexDocument(ctx, IndexRequest{
		OrgID:      orgID,
		SourceType: "issue",
		SourceID:   issueID,
		Content:    content,
	})
	if err != nil {
		log.Printf("Warning: failed to index issue %s: %v", issueID, err)
	}
}

func (i *Indexer) IndexComment(ctx context.Context, orgID, commentID uuid.UUID, content string) {
	if i == nil || i.service == nil {
		return
	}

	if content == "" {
		return
	}

	err := i.service.IndexDocument(ctx, IndexRequest{
		OrgID:      orgID,
		SourceType: "comment",
		SourceID:   commentID,
		Content:    fmt.Sprintf("Comment: %s", content),
	})
	if err != nil {
		log.Printf("Warning: failed to index comment %s: %v", commentID, err)
	}
}

func (i *Indexer) DeleteTask(ctx context.Context, orgID, taskID uuid.UUID) {
	if i == nil || i.service == nil {
		return
	}

	err := i.service.DeleteDocument(ctx, orgID, "task", taskID)
	if err != nil {
		log.Printf("Warning: failed to delete task index %s: %v", taskID, err)
	}
}

func (i *Indexer) DeleteIssue(ctx context.Context, orgID, issueID uuid.UUID) {
	if i == nil || i.service == nil {
		return
	}

	err := i.service.DeleteDocument(ctx, orgID, "issue", issueID)
	if err != nil {
		log.Printf("Warning: failed to delete issue index %s: %v", issueID, err)
	}
}

func (i *Indexer) DeleteComment(ctx context.Context, orgID, commentID uuid.UUID) {
	if i == nil || i.service == nil {
		return
	}

	err := i.service.DeleteDocument(ctx, orgID, "comment", commentID)
	if err != nil {
		log.Printf("Warning: failed to delete comment index %s: %v", commentID, err)
	}
}

func (i *Indexer) IndexDocument(ctx context.Context, orgID, documentID uuid.UUID, title, content string) error {
	if i == nil || i.service == nil {
		return nil
	}

	if content == "" {
		return nil
	}

	return i.service.IndexDocument(ctx, IndexRequest{
		OrgID:      orgID,
		SourceType: "document",
		SourceID:   documentID,
		Content:    fmt.Sprintf("Document: %s\n\n%s", title, content),
	})
}

func (i *Indexer) DeleteDocument(ctx context.Context, orgID, documentID uuid.UUID) {
	if i == nil || i.service == nil {
		return
	}

	err := i.service.DeleteDocument(ctx, orgID, "document", documentID)
	if err != nil {
		log.Printf("Warning: failed to delete document index %s: %v", documentID, err)
	}
}

func (i *Indexer) IndexTaskDocument(ctx context.Context, orgID, taskID uuid.UUID, filename, content string) {
	if i == nil || i.service == nil {
		return
	}

	if content == "" {
		return
	}

	docContent := fmt.Sprintf("Task Document: %s\nContent: %s", filename, content)
	err := i.service.IndexDocument(ctx, IndexRequest{
		OrgID:      orgID,
		SourceType: "task_document",
		SourceID:   taskID,
		Content:    docContent,
	})
	if err != nil {
		log.Printf("Warning: failed to index task document %s: %v", taskID, err)
	}
}

package ai

import (
	"fmt"
	"strings"
)

type PromptTemplate struct {
	systemRole   string
	instructions []string
	constraints  []string
}

func NewPromptTemplate(role string) *PromptTemplate {
	return &PromptTemplate{
		systemRole:   role,
		instructions: make([]string, 0),
		constraints:  make([]string, 0),
	}
}

func (p *PromptTemplate) AddInstruction(instruction string) *PromptTemplate {
	p.instructions = append(p.instructions, instruction)
	return p
}

func (p *PromptTemplate) AddConstraint(constraint string) *PromptTemplate {
	p.constraints = append(p.constraints, constraint)
	return p
}

func (p *PromptTemplate) Build(context, query string) string {
	var sb strings.Builder

	sb.WriteString("ROLE: ")
	sb.WriteString(p.systemRole)
	sb.WriteString("\n\n")

	if len(p.instructions) > 0 {
		sb.WriteString("INSTRUCTIONS:\n")
		for i, inst := range p.instructions {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, inst))
		}
		sb.WriteString("\n")
	}

	if len(p.constraints) > 0 {
		sb.WriteString("CONSTRAINTS:\n")
		for i, cons := range p.constraints {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, cons))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("CONTEXT:\n")
	sb.WriteString(context)
	sb.WriteString("\n\n")

	sb.WriteString("QUERY: ")
	sb.WriteString(query)
	sb.WriteString("\n\n")

	sb.WriteString("RESPONSE:")

	return sb.String()
}

func DocumentVerificationPrompt(taskTitle, taskDescription, documentContent string) string {
	template := NewPromptTemplate("Document Verification Assistant")
	template.AddInstruction("Analyze the submitted document against the assigned task requirements")
	template.AddInstruction("Identify what work was completed based ONLY on the document content")
	template.AddInstruction("Verify if the completed work matches the task requirements")
	template.AddInstruction("Highlight any missing deliverables or unclear sections")
	template.AddConstraint("You MUST answer only from the provided document content")
	template.AddConstraint("Do NOT assume work was done if not explicitly shown in the document")
	template.AddConstraint("Do NOT use external knowledge")
	template.AddConstraint("If evidence is insufficient, state this clearly")

	context := fmt.Sprintf(`TASK ASSIGNED:
Title: %s
Description: %s

SUBMITTED DOCUMENT CONTENT:
%s`, taskTitle, taskDescription, documentContent)

	query := "Does this document demonstrate completion of the assigned task? What work was done and what is missing?"

	return template.Build(context, query)
}

func TaskSummaryPrompt(taskTitle, documentContent string) string {
	template := NewPromptTemplate("Task Completion Summarizer")
	template.AddInstruction("Summarize what work was completed based on the submitted document")
	template.AddInstruction("Focus on main deliverables and key outcomes")
	template.AddInstruction("Keep the summary concise (3-5 sentences)")
	template.AddInstruction("Highlight any notable findings or results")
	template.AddConstraint("Base your summary ONLY on the document content provided")
	template.AddConstraint("Do NOT infer work that is not explicitly described")
	template.AddConstraint("Do NOT add recommendations or speculation")

	context := fmt.Sprintf(`TASK: %s

DOCUMENT CONTENT:
%s`, taskTitle, documentContent)

	query := "Provide a concise summary of the completed work for manager review"

	return template.Build(context, query)
}

func RAGQueryPrompt(retrievedDocs []string, userQuery string) string {
	template := NewPromptTemplate("Knowledge Assistant")
	template.AddInstruction("Answer the user's question using ONLY the retrieved documents")
	template.AddInstruction("Cite specific information from the context when possible")
	template.AddInstruction("If the context does not contain enough information, state this clearly")
	template.AddConstraint("You MUST NOT use external knowledge")
	template.AddConstraint("You MUST NOT make assumptions")
	template.AddConstraint("If you cannot answer from context, say: 'I cannot answer this based on the available documents'")

	context := strings.Join(retrievedDocs, "\n\n---\n\n")
	if context == "" {
		context = "[No relevant documents found]"
	}

	return template.Build(context, userQuery)
}

func AdminReportPrompt(taskData string) string {
	template := NewPromptTemplate("Operations Assistant")
	template.AddInstruction("Analyze the organization's task data and create a structured report")
	template.AddInstruction("Use Markdown format with headings (##, ###) and numbered lists only")
	template.AddInstruction("Do not use asterisks, dashes, or bullet points")
	template.AddInstruction("Use minimal punctuation - avoid exclamation marks, ellipses, emojis")
	template.AddInstruction("Keep sentences short and clear")
	template.AddConstraint("Base analysis ONLY on provided task data")
	template.AddConstraint("Do NOT speculate about missing information")

	context := fmt.Sprintf(`TASK DATA:
%s`, taskData)

	query := `Generate a report with these sections:
## Overall Status (2-4 sentences summarizing current state)
## Key Risks (numbered list, up to 5 items)
## Recommended Next Actions (numbered list, up to 5 items)
## Task Stats (show counts and metrics)`

	return template.Build(context, query)
}

func IssueSummaryPrompt(issueTitle, issueDescription string) string {
	template := NewPromptTemplate("Issue Analysis Assistant")
	template.AddInstruction("Analyze the issue description and provide a structured summary")
	template.AddInstruction("Identify the core problem, impact, and suggested priority")
	template.AddInstruction("Keep summary concise and actionable")
	template.AddConstraint("Base analysis ONLY on the issue description provided")
	template.AddConstraint("Do NOT add technical solutions or implementation details")

	context := fmt.Sprintf(`ISSUE TITLE: %s

ISSUE DESCRIPTION:
%s`, issueTitle, issueDescription)

	query := "Provide a structured summary: Problem, Impact, Suggested Priority"

	return template.Build(context, query)
}

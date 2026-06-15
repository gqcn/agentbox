// This file implements Go text/template parsing, variable reference
// validation, sample variable preparation, and rendering for prompt templates.

package prompt

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"text/template/parse"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
)

// validateTemplateContent parses a template and verifies variable governance rules.
func validateTemplateContent(def templateDefinition, content string, requireRequiredReferences bool) error {
	tmpl, err := parsePromptTemplate(def.Code, content)
	if err != nil {
		return err
	}
	references := referencedTemplateVariables(tmpl)
	declared := declaredVariables(def)
	for name := range references {
		if _, ok := declared[name]; !ok {
			return bizerr.WrapCode(gerror.Newf("unknown prompt template variable %q", name), CodePromptInvalidInput)
		}
	}
	if requireRequiredReferences {
		for _, variable := range def.Variables {
			if variable.Required {
				if _, ok := references[variable.Name]; !ok {
					return bizerr.WrapCode(gerror.Newf("required prompt template variable %q is missing", variable.Name), CodePromptInvalidInput)
				}
			}
		}
	}
	if _, err := executePromptTemplate(tmpl, sampleVariables(def)); err != nil {
		return err
	}
	return nil
}

// renderTemplateContent renders a template using runtime or preview variables.
func renderTemplateContent(def templateDefinition, content string, variables map[string]string, requireRequiredValues bool) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", bizerr.NewCode(CodePromptInvalidInput)
	}
	tmpl, err := parsePromptTemplate(def.Code, content)
	if err != nil {
		return "", err
	}
	references := referencedTemplateVariables(tmpl)
	declared := declaredVariables(def)
	for name := range references {
		if _, ok := declared[name]; !ok {
			return "", bizerr.WrapCode(gerror.Newf("unknown prompt template variable %q", name), CodePromptInvalidInput)
		}
	}
	if requireRequiredValues {
		for _, variable := range def.Variables {
			if !variable.Required {
				continue
			}
			if strings.TrimSpace(variables[variable.Name]) == "" {
				return "", bizerr.WrapCode(gerror.Newf("required prompt template variable %q is missing", variable.Name), CodePromptInvalidInput)
			}
		}
	}
	return executePromptTemplate(tmpl, variables)
}

// parsePromptTemplate parses content with strict missing-key behavior.
func parsePromptTemplate(code string, content string) (*template.Template, error) {
	tmpl, err := template.New(code).Option("missingkey=error").Parse(content)
	if err != nil {
		return nil, bizerr.WrapCode(gerror.Wrap(err, "parse prompt template"), CodePromptInvalidInput)
	}
	return tmpl, nil
}

// executePromptTemplate runs a parsed template and returns its rendered text.
func executePromptTemplate(tmpl *template.Template, variables map[string]string) (string, error) {
	var output bytes.Buffer
	if err := tmpl.Execute(&output, variables); err != nil {
		return "", bizerr.WrapCode(gerror.Wrap(err, "render prompt template"), CodePromptInvalidInput)
	}
	return output.String(), nil
}

// declaredVariables maps variable names declared by a template definition.
func declaredVariables(def templateDefinition) map[string]struct{} {
	declared := make(map[string]struct{}, len(def.Variables))
	for _, variable := range def.Variables {
		declared[variable.Name] = struct{}{}
	}
	return declared
}

// sampleVariables returns safe sample values for every declared variable.
func sampleVariables(def templateDefinition) map[string]string {
	values := make(map[string]string, len(def.Variables))
	for _, variable := range def.Variables {
		values[variable.Name] = variable.SampleValue
	}
	return values
}

// previewVariables merges caller-provided preview values with safe samples.
func previewVariables(def templateDefinition, supplied map[string]string) map[string]string {
	values := sampleVariables(def)
	for key, value := range supplied {
		values[key] = value
	}
	return values
}

// referencedTemplateVariables collects top-level data-field references.
func referencedTemplateVariables(tmpl *template.Template) map[string]struct{} {
	references := map[string]struct{}{}
	if tmpl == nil || tmpl.Tree == nil || tmpl.Tree.Root == nil {
		return references
	}
	collectNodeVariables(tmpl.Tree.Root, references)
	return references
}

// collectNodeVariables recursively walks text/template parse nodes.
func collectNodeVariables(node parse.Node, references map[string]struct{}) {
	if node == nil {
		return
	}
	switch current := node.(type) {
	case *parse.ListNode:
		if current == nil {
			return
		}
		for _, child := range current.Nodes {
			collectNodeVariables(child, references)
		}
	case *parse.ActionNode:
		if current == nil {
			return
		}
		collectPipeVariables(current.Pipe, references)
	case *parse.IfNode:
		if current == nil {
			return
		}
		collectPipeVariables(current.Pipe, references)
		collectNodeVariables(current.List, references)
		collectNodeVariables(current.ElseList, references)
	case *parse.RangeNode:
		if current == nil {
			return
		}
		collectPipeVariables(current.Pipe, references)
		collectNodeVariables(current.List, references)
		collectNodeVariables(current.ElseList, references)
	case *parse.WithNode:
		if current == nil {
			return
		}
		collectPipeVariables(current.Pipe, references)
		collectNodeVariables(current.List, references)
		collectNodeVariables(current.ElseList, references)
	case *parse.TemplateNode:
		if current == nil {
			return
		}
		collectPipeVariables(current.Pipe, references)
	case *parse.PipeNode:
		if current == nil {
			return
		}
		collectPipeVariables(current, references)
	case *parse.CommandNode:
		if current == nil {
			return
		}
		collectCommandVariables(current, references)
	case *parse.FieldNode:
		if current == nil {
			return
		}
		addFieldReference(current.Ident, references)
	case *parse.ChainNode:
		if current == nil {
			return
		}
		collectNodeVariables(current.Node, references)
		addFieldReference(current.Field, references)
	case *parse.VariableNode:
		if current == nil {
			return
		}
		if len(current.Ident) > 1 && current.Ident[0] == "$" {
			references[current.Ident[1]] = struct{}{}
		}
	}
}

// collectPipeVariables walks all commands inside a template pipeline.
func collectPipeVariables(pipe *parse.PipeNode, references map[string]struct{}) {
	if pipe == nil {
		return
	}
	for _, command := range pipe.Cmds {
		collectCommandVariables(command, references)
	}
}

// collectCommandVariables walks command arguments inside a pipeline.
func collectCommandVariables(command *parse.CommandNode, references map[string]struct{}) {
	if command == nil {
		return
	}
	for _, arg := range command.Args {
		collectNodeVariables(arg, references)
	}
}

// addFieldReference records the top-level data field from a field chain.
func addFieldReference(fields []string, references map[string]struct{}) {
	if len(fields) == 0 {
		return
	}
	name := strings.TrimSpace(fields[0])
	if name == "" {
		return
	}
	if strings.HasPrefix(name, "$") {
		return
	}
	references[name] = struct{}{}
}

// formatVariableNames produces a deterministic debug string for tests.
func formatVariableNames(values map[string]struct{}) string {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	return fmt.Sprintf("%v", names)
}

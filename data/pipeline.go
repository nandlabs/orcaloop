package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidType = errors.New("invalid type")
var ErrKeyNotFound = errors.New("key not found")

// Pipeline is a struct that represents a workflow context

const (
	InstanceIdKey = "__instanceId__"
	StepIdKey     = "__stepId__"
	ErrorKey      = "__error__"
)

// Pipeline represents a pipeline that processes data stored in a map.
// The data is stored as key-value pairs where the key is a string and the value can be of any type.
type Pipeline struct {
	data map[string]any
}

// ExtractValue retrieves a value of type T from the Pipeline using the provided key.
// If the value is not of type T, it returns an ErrInvalidType error.
//
// Parameters:
//   - c: A pointer to the Pipeline from which to extract the value.
//   - key: The key associated with the value to be retrieved.
//
// Returns:
//   - value: The value of type T associated with the provided key.
//   - err: An error if the key does not exist or the value is not of type T.
func ExtractValue[T any](c *Pipeline, key string) (value T, err error) {

	var v any
	v, err = c.Get(key)
	if err != nil {
		return
	}
	if _, ok := v.(T); ok {
		value = v.(T)
	} else {
		err = ErrInvalidType
	}
	return
}

// Wrap initializes a new Pipeline with the provided ID and values.
// It sets the InstanceIdKey in the pipeline with the given ID.
//
// Parameters:
//   - id: A string representing the unique identifier for the pipeline instance.
//   - values: A map containing key-value pairs to initialize the pipeline data.
//
// Returns:
//   - pipeline: A Pipeline interface initialized with the provided values and ID.
func Wrap(id string, values map[string]any) (pipeline *Pipeline) {
	pipeline = &Pipeline{
		data: values,
	}
	pipeline.Set(InstanceIdKey, id)
	return
}

// NewPipeline creates a new instance of a Pipeline with the given ID.
// It initializes the pipeline's data map and sets the InstanceIdKey to the provided ID.
//
// Parameters:
//   - id: A string representing the unique identifier for the pipeline instance.
//
// Returns:
//   - pipeline: A Pipeline instance with the specified ID.
func NewPipeline(id string) (pipeline *Pipeline) {
	pipeline = &Pipeline{
		data: make(map[string]any),
	}
	pipeline.Set(InstanceIdKey, id)

	return
}

// NewPipelineFrom creates a new instance of a Pipeline with the given ID and initial values.
// It initializes the pipeline's data map and sets the provided values.
// Additionally, it sets the InstanceIdKey to the provided ID.
//
// Parameters:
//   - id: A string representing the unique identifier for the pipeline instance.
//   - values: A map containing initial key-value pairs to be set in the pipeline.
//
// Returns:
//
//	A Pipeline instance with the specified ID and initial values.
func NewPipelineFrom(id string, values map[string]any) (pipeline *Pipeline) {
	pipeline = &Pipeline{

		data: make(map[string]any),
	}
	for k, v := range values {
		pipeline.Set(k, v)
	}
	pipeline.Set(InstanceIdKey, id)
	return
}

// Id returns the instance ID of the Pipeline.
// It extracts the ID value from the Pipeline using the InstanceIdKey.
func (p *Pipeline) Id() (id string) {
	id, _ = ExtractValue[string](p, InstanceIdKey)
	return
}

// // StepId retrieves the step identifier from the Pipeline instance.
// // It uses the ExtractValue function to obtain the value associated with the StepIdKey.
// // Returns the step identifier as a string.
// func (p *Pipeline) StepId() (stepId string) {
// 	stepId, _ = ExtractValue[string](p, StepIdKey)
// 	return
// }

// Get retrieves the value associated with the given key from the Pipeline.
// If the key is found, the value is returned along with a nil error.
// If the key is not found, an ErrKeyNotFound error is returned.
//
// Parameters:
//   - key: The key to look up in the Pipeline.
//
// Returns:
//   - value: The value associated with the key, if found.
//   - err: An error indicating whether the key was found or not.
func (p *Pipeline) Get(key string) (value any, err error) {

	if v, ok := p.data[key]; ok {
		value = v
	} else {
		err = ErrKeyNotFound
	}
	return
}

// Has checks if the given key exists in the Pipeline's data map.
// It returns true if the key is present, otherwise false.
//
// Parameters:
//
//	key - the key to be checked in the data map.
//
// Returns:
//
//	bool - true if the key exists, false otherwise.
func (p *Pipeline) Has(key string) bool {
	_, ok := p.data[key]
	return ok
}

// Keys returns a slice of all the keys present in the Pipeline's data.
// It iterates over the map and collects each key into a slice, which is then returned.
func (p *Pipeline) Keys() []string {

	keys := make([]string, 0, len(p.data))
	for k := range p.data {
		keys = append(keys, k)
	}
	return keys
}

// Set assigns the given value to the specified key in the Pipeline's data map.
// If the key already exists, its value will be updated.
//
// Parameters:
//
//	key: The key to which the value should be assigned.
//	value: The value to be assigned to the specified key.
//
// Returns:
//
//	An error if the operation fails, otherwise nil.
func (p *Pipeline) Set(key string, value any) error {

	p.data[key] = value
	return nil
}

// Delete removes the entry with the specified key from the Pipeline's data.
// If the key does not exist, the function does nothing and returns nil.
//
// Parameters:
//
//	key - The key of the entry to be deleted.
//
// Returns:
//
//	An error if the deletion fails, otherwise nil.
func (p *Pipeline) Delete(key string) error {
	delete(p.data, key)
	return nil
}

// Map creates and returns a new map with the same key-value pairs as the
// Pipeline's internal data. The returned map has keys of type string and
// values of type any.
func (p *Pipeline) Map() map[string]any {
	data := make(map[string]any, len(p.data))
	for k, v := range p.data {
		data[k] = v
	}
	return data
}

// GetError retrieves the error value from the Pipeline instance.
// It uses the ExtractValue function to obtain the value associated with the ErrorKey.
// Returns the error value as an error.
func (p *Pipeline) GetError() (errMsg string) {
	errMsg, _ = ExtractValue[string](p, ErrorKey)
	return
}

// SetError assigns the provided error message to the Pipeline instance.
// It uses the Set function to assign the error message to the ErrorKey.
//
// Parameters:
//   - errMsg: A string representing the error message to be assigned.
func (p *Pipeline) SetError(errMsg string) {
	p.Set(ErrorKey, errMsg)
}

// MergeFrom combines the key-value pairs from the provided map into the Pipeline's data.
// If a key already exists in the Pipeline, its value will be updated with the new value.
//
// Parameters:
//   - data: A map containing key-value pairs to be merged into the Pipeline.
//
// Returns:
//
//	An error if the merge operation fails, otherwise nil.
func (p *Pipeline) MergeFrom(data map[string]any) error {
	for k, v := range data {
		p.data[k] = v
	}
	return nil
}

// Merge combines the key-value pairs from the provided Pipeline into the current Pipeline.
// If a key already exists in the current Pipeline, its value will be updated with the new value.
// Parameters:
//   - pipeline: A Pipeline instance containing key-value pairs to be merged into the current Pipeline.
//
// Returns:
//
//	An error if the merge operation fails, otherwise nil.
func (p *Pipeline) Merge(pipeline *Pipeline) error {
	for k, v := range pipeline.data {
		p.data[k] = v
	}
	return nil
}

// Clone creates a deep copy of the current Pipeline instance.
// It returns a new Pipeline instance with a duplicated map containing
// the same key-value pairs as the original Pipeline.
func (p *Pipeline) Clone() *Pipeline {
	data := make(map[string]any, len(p.data))
	for k, v := range p.data {
		data[k] = v
	}
	return &Pipeline{
		data: data,
	}
}

// evaluateCondition evaluates a condition string using variables from the context.
// It supports basic comparison and logical operators: ==, !=, <, >, <=, >=, &&, ||.
func (p *Pipeline) EvaluateCondition(condition string) (bool, error) {
	// Tokenize the input condition string into smaller parts (e.g., variables, operators).
	tokens := tokenize(condition)

	// Convert the infix tokens (e.g., "a == b && c") into postfix notation (Reverse Polish Notation).
	postfix, err := infixToPostfix(tokens)
	if err != nil {
		return false, err
	}

	// Evaluate the postfix expression using the workflow context.
	return evaluatePostfix(postfix, p)
}

// tokenize splits the condition string into tokens for parsing.
func tokenize(condition string) []string {
	var tokens []string
	var currentToken strings.Builder

	// Check if a character is part of an operator
	isOperator := func(r rune) bool {
		return strings.ContainsRune("!=<>&|()", r)
	}

	// Iterate through each character in the condition
	for _, ch := range condition {
		switch {
		case ch == ' ': // Skip spaces
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
		case isOperator(ch): // If the character is an operator, split the token
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, string(ch))
		default: // Add characters to the current token
			currentToken.WriteRune(ch)
		}
	}
	// Append the last token, if any
	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}
	return tokens
}

// precedence determines the precedence of operators.
func precedence(op string) int {
	switch op {
	case "||": // Logical OR has the lowest precedence
		return 1
	case "&&": // Logical AND has higher precedence than OR
		return 2
	case "==", "!=", "<", ">", "<=", ">=": // Comparison operators
		return 3
	case "(": // Parentheses have no precedence themselves
		return 0
	default: // Unknown operators
		return -1
	}
}

// infixToPostfix converts an infix expression (e.g., "a == b && c") to postfix (Reverse Polish Notation).
func infixToPostfix(tokens []string) ([]string, error) {
	var postfix []string
	var stack []string

	// Process each token
	for _, token := range tokens {
		switch token {
		case "&&", "||", "==", "!=", "<", ">", "<=", ">=": // If the token is an operator
			// Pop higher or equal precedence operators from the stack to postfix
			for len(stack) > 0 && precedence(stack[len(stack)-1]) >= precedence(token) {
				postfix = append(postfix, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token) // Push the current operator onto the stack
		case "(": // Push opening parentheses onto the stack
			stack = append(stack, token)
		case ")": // Process closing parentheses
			// Pop from the stack until an opening parenthesis is encountered
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				postfix = append(postfix, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			// Check for mismatched parentheses
			if len(stack) == 0 {
				return nil, errors.New("mismatched parentheses")
			}
			stack = stack[:len(stack)-1] // Pop the opening parenthesis
		default: // Otherwise, it's an operand (variable or literal)
			postfix = append(postfix, token)
		}
	}

	// Pop any remaining operators from the stack
	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return nil, errors.New("mismatched parentheses")
		}
		postfix = append(postfix, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return postfix, nil
}

// evaluatePostfix evaluates a postfix expression using the context.
func evaluatePostfix(postfix []string, pipeline *Pipeline) (bool, error) {
	var stack []any

	// Helper function to compare two operands with an operator
	compare := func(a, b any, op string) (bool, error) {
		switch op {
		case "==": // Equality
			return a == b, nil
		case "!=": // Inequality
			return a != b, nil
		case "<": // Less than
			return a.(float64) < b.(float64), nil
		case ">": // Greater than
			return a.(float64) > b.(float64), nil
		case "<=": // Less than or equal
			return a.(float64) <= b.(float64), nil
		case ">=": // Greater than or equal
			return a.(float64) >= b.(float64), nil
		default:
			return false, fmt.Errorf("unknown operator: %s", op)
		}
	}

	// Process each token in the postfix expression
	for _, token := range postfix {
		switch token {
		case "&&", "||": // Logical operators
			if len(stack) < 2 {
				return false, errors.New("invalid logical expression")
			}
			// Pop two operands
			b := stack[len(stack)-1].(bool)
			a := stack[len(stack)-2].(bool)
			stack = stack[:len(stack)-2]

			// Perform the logical operation
			if token == "&&" {
				stack = append(stack, a && b)
			} else {
				stack = append(stack, a || b)
			}
		case "==", "!=", "<", ">", "<=", ">=": // Comparison operators
			if len(stack) < 2 {
				return false, errors.New("invalid comparison expression")
			}
			// Pop two operands
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			// Perform the comparison
			result, err := compare(a, b, token)
			if err != nil {
				return false, err
			}
			stack = append(stack, result)
		default: // Operands (variables or literals)
			value, err := parseToken(token, pipeline)
			if err != nil {
				return false, err
			}
			stack = append(stack, value)
		}
	}

	// There should be exactly one result left on the stack
	if len(stack) != 1 {
		return false, errors.New("invalid postfix expression")
	}
	return stack[0].(bool), nil
}

// parseToken converts a token into a value using the context.
func parseToken(token string, pipeline *Pipeline) (any, error) {
	// If the token is a string literal (surrounded by quotes)
	if strings.HasPrefix(token, `"`) && strings.HasSuffix(token, `"`) {
		return token[1 : len(token)-1], nil
	}

	// If the token is a number, parse it
	if value, err := strconv.ParseFloat(token, 64); err == nil {
		return value, nil
	}

	// Otherwise, assume it's a variable in the context
	if pipeline.Has(token) {
		return pipeline.Get(token)
	}

	return nil, fmt.Errorf("unknown variable or value: %s", token)
}

package execmodel

import "fmt"

type ResourceMode string

const (
	ResourceModeFinite   ResourceMode = "finite"
	ResourceModeInfinite ResourceMode = "infinite"
)

func (m ResourceMode) String() string {
	return string(m)
}

type FeedbackMode string

const (
	FeedbackModeEnabled  FeedbackMode = "enabled"
	FeedbackModeDisabled FeedbackMode = "disabled"
)

func (m FeedbackMode) String() string {
	return string(m)
}

type Model struct {
	Resource ResourceMode
	Feedback FeedbackMode
}

func (m Model) String() string {
	return fmt.Sprintf("resource=%s feedback=%s", m.Resource, m.Feedback)
}

func (m Model) Validate() error {
	switch m.Resource {
	case ResourceModeFinite, ResourceModeInfinite:
	default:
		return fmt.Errorf("execmodel.Model.Validate: unknown resource mode: %q", m.Resource)
	}
	switch m.Feedback {
	case FeedbackModeEnabled, FeedbackModeDisabled:
	default:
		return fmt.Errorf("execmodel.Model.Validate: unknown feedback mode: %q", m.Feedback)
	}
	return nil
}

func DefaultModel() Model {
	return Model{
		Resource: ResourceModeInfinite,
		Feedback: FeedbackModeDisabled,
	}
}

func ParseResourceMode(s string) (ResourceMode, error) {
	switch s {
	case "":
		return DefaultModel().Resource, nil
	case ResourceModeFinite.String():
		return ResourceModeFinite, nil
	case ResourceModeInfinite.String():
		return ResourceModeInfinite, nil
	}

	return ResourceMode(""), fmt.Errorf("execmodel.ParseResourceMode: unknown resource mode: %q", s)
}

func ParseFeedbackMode(s string) (FeedbackMode, error) {
	switch s {
	case "":
		return DefaultModel().Feedback, nil
	case FeedbackModeEnabled.String():
		return FeedbackModeEnabled, nil
	case FeedbackModeDisabled.String():
		return FeedbackModeDisabled, nil
	}

	return FeedbackMode(""), fmt.Errorf("execmodel.ParseFeedbackMode: unknown feedback mode: %q", s)
}

func ParseModel(resourceMode string, feedbackMode string) (Model, error) {
	r, err := ParseResourceMode(resourceMode)
	if err != nil {
		return Model{}, fmt.Errorf("execmodel.ParseModel: %w", err)
	}
	f, err := ParseFeedbackMode(feedbackMode)
	if err != nil {
		return Model{}, fmt.Errorf("execmodel.ParseModel: %w", err)
	}
	return Model{Resource: r, Feedback: f}, nil
}

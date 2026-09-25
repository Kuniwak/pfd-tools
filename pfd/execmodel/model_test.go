package execmodel_test

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
)

func TestParseResourceMode(t *testing.T) {
	testCases := map[string]struct {
		Input    string
		Expected execmodel.ResourceMode
		HasError bool
	}{
		"finite":                {Input: "finite", Expected: execmodel.ResourceModeFinite},
		"infinite":              {Input: "infinite", Expected: execmodel.ResourceModeInfinite},
		"empty means default":   {Input: "", Expected: execmodel.DefaultModel().Resource},
		"unknown is an error":   {Input: "unlimited", HasError: true},
		"case is not converted": {Input: "Finite", HasError: true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actual, err := execmodel.ParseResourceMode(tc.Input)
			if tc.HasError {
				if err == nil {
					t.Errorf("expected error for %q, but got none (result: %q)", tc.Input, actual)
				}

				if actual != "" {
					t.Errorf("expected the zero value on error, but got %q", actual)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.Input, err)
			}
			if actual != tc.Expected {
				t.Errorf("expected %q, but got %q", tc.Expected, actual)
			}
		})
	}
}

func TestParseFeedbackMode(t *testing.T) {
	testCases := map[string]struct {
		Input    string
		Expected execmodel.FeedbackMode
		HasError bool
	}{
		"enabled":             {Input: "enabled", Expected: execmodel.FeedbackModeEnabled},
		"disabled":            {Input: "disabled", Expected: execmodel.FeedbackModeDisabled},
		"empty means default": {Input: "", Expected: execmodel.DefaultModel().Feedback},
		"unknown is an error": {Input: "on", HasError: true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actual, err := execmodel.ParseFeedbackMode(tc.Input)
			if tc.HasError {
				if err == nil {
					t.Errorf("expected error for %q, but got none (result: %q)", tc.Input, actual)
				}

				if actual != "" {
					t.Errorf("expected the zero value on error, but got %q", actual)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.Input, err)
			}
			if actual != tc.Expected {
				t.Errorf("expected %q, but got %q", tc.Expected, actual)
			}
		})
	}
}

func TestDefaultModelIsInfiniteAndFeedbackDisabled(t *testing.T) {
	expected := execmodel.Model{
		Resource: execmodel.ResourceModeInfinite,
		Feedback: execmodel.FeedbackModeDisabled,
	}
	if actual := execmodel.DefaultModel(); actual != expected {
		t.Errorf("expected %+v, but got %+v", expected, actual)
	}
}

func TestParseModel(t *testing.T) {
	testCases := map[string]struct {
		ResourceMode string
		FeedbackMode string
		Expected     execmodel.Model
		HasError     bool
	}{
		"both empty means default": {
			ResourceMode: "", FeedbackMode: "",
			Expected: execmodel.DefaultModel(),
		},
		"finite and enabled": {
			ResourceMode: "finite", FeedbackMode: "enabled",
			Expected: execmodel.Model{Resource: execmodel.ResourceModeFinite, Feedback: execmodel.FeedbackModeEnabled},
		},
		"infinite and enabled": {
			ResourceMode: "infinite", FeedbackMode: "enabled",
			Expected: execmodel.Model{Resource: execmodel.ResourceModeInfinite, Feedback: execmodel.FeedbackModeEnabled},
		},
		"unknown resource mode is an error": {ResourceMode: "none", FeedbackMode: "enabled", HasError: true},
		"unknown feedback mode is an error": {ResourceMode: "finite", FeedbackMode: "yes", HasError: true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actual, err := execmodel.ParseModel(tc.ResourceMode, tc.FeedbackMode)
			if tc.HasError {
				if err == nil {
					t.Errorf("expected error for (%q, %q), but got none (result: %+v)", tc.ResourceMode, tc.FeedbackMode, actual)
				}
				if actual != (execmodel.Model{}) {
					t.Errorf("expected the zero value on error, but got %+v", actual)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for (%q, %q): %v", tc.ResourceMode, tc.FeedbackMode, err)
			}
			if actual != tc.Expected {
				t.Errorf("expected %+v, but got %+v", tc.Expected, actual)
			}
		})
	}
}

func TestModelString(t *testing.T) {
	testCases := map[string]struct {
		Input    execmodel.Model
		Expected string
	}{
		"infinite and disabled": {
			Input:    execmodel.Model{Resource: execmodel.ResourceModeInfinite, Feedback: execmodel.FeedbackModeDisabled},
			Expected: "resource=infinite feedback=disabled",
		},
		"finite and enabled": {
			Input:    execmodel.Model{Resource: execmodel.ResourceModeFinite, Feedback: execmodel.FeedbackModeEnabled},
			Expected: "resource=finite feedback=enabled",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if actual := tc.Input.String(); actual != tc.Expected {
				t.Errorf("expected %q, but got %q", tc.Expected, actual)
			}
		})
	}
}

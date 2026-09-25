package cmd

import "testing"

func TestParseMode(t *testing.T) {
	testCases := map[string]struct {
		OutputFormat string
		Inplace      bool
		PlanPath     string
		MinID        string
		Want         Mode
		WantErr      bool
	}{
		"no flags decides and applies a plan": {
			OutputFormat: OutputFormatDrawio,
			Want:         ModeRenumber,
		},
		"-inplace alone decides and applies a plan": {
			OutputFormat: OutputFormatDrawio,
			Inplace:      true,
			Want:         ModeRenumber,
		},
		"-o tsv emits a plan": {
			OutputFormat: OutputFormatTSV,
			Want:         ModeEmitPlan,
		},
		"-renum-plan applies the given plan": {
			OutputFormat: OutputFormatDrawio,
			PlanPath:     "renum.tsv",
			Want:         ModeApplyPlan,
		},
		"-renum-plan with -inplace applies the given plan": {
			OutputFormat: OutputFormatDrawio,
			Inplace:      true,
			PlanPath:     "renum.tsv",
			Want:         ModeApplyPlan,
		},
		"-o tsv with -inplace is an error": {
			OutputFormat: OutputFormatTSV,
			Inplace:      true,
			WantErr:      true,
		},
		"-o tsv with -renum-plan extends the given plan": {
			OutputFormat: OutputFormatTSV,
			PlanPath:     "renum.tsv",
			Want:         ModeExtendPlan,
		},
		"-o tsv with -renum-plan and -inplace is an error": {
			OutputFormat: OutputFormatTSV,
			Inplace:      true,
			PlanPath:     "renum.tsv",
			WantErr:      true,
		},
		"-o maxid emits the max IDs already used": {
			OutputFormat: OutputFormatMaxID,
			Want:         ModeEmitMaxID,
		},
		"-o maxid with -inplace is an error": {
			OutputFormat: OutputFormatMaxID,
			Inplace:      true,
			WantErr:      true,
		},
		"-o maxid with -renum-plan is an error": {
			OutputFormat: OutputFormatMaxID,
			PlanPath:     "renum.tsv",
			WantErr:      true,
		},
		"-min-id decides and applies a plan": {
			OutputFormat: OutputFormatDrawio,
			MinID:        "P12",
			Want:         ModeRenumber,
		},
		"-min-id with -o tsv emits a plan": {
			OutputFormat: OutputFormatTSV,
			MinID:        "P12",
			Want:         ModeEmitPlan,
		},
		"-min-id with -renum-plan is an error": {
			OutputFormat: OutputFormatDrawio,
			PlanPath:     "renum.tsv",
			MinID:        "P12",
			WantErr:      true,
		},
		"-min-id with -o maxid is an error": {
			OutputFormat: OutputFormatMaxID,
			MinID:        "P12",
			WantErr:      true,
		},
		"an unknown output format is an error": {
			OutputFormat: "json",
			WantErr:      true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := ParseMode(tc.OutputFormat, tc.Inplace, tc.PlanPath, tc.MinID)
			if tc.WantErr {
				if err == nil {
					t.Fatalf("ParseMode: err = nil, want an error (got mode %q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseMode: %v", err)
			}
			if got != tc.Want {
				t.Errorf("ParseMode = %q, want %q", got, tc.Want)
			}
		})
	}
}

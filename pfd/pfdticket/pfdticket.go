package pfdticket

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/Kuniwak/pfd-tools/locale"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable"
	"github.com/Kuniwak/pfd-tools/sets"
)

const (
	PlaceholderDeliverableLinkJa = "（完了時に URL を記入）"
	PlaceholderDeliverableLinkEn = "(fill in the URL when done)"
)

const (
	LocationProducedTemplateJa = "所在: %s のチケット"
	LocationProducedTemplateEn = "Location: %s ticket"
	LocationUnknownJa          = "成果物所在不明"
	LocationUnknownEn          = "deliverable location unknown"
)

var (
	FormatColumn         = pfdtable.ColumnAliases{Ja: "フォーマット", En: "Format"}
	ReviewCriteriaColumn = pfdtable.ColumnAliases{Ja: "レビュー基準", En: "Review Criteria"}
	ReviewerColumn       = pfdtable.ColumnAliases{Ja: "レビューア", En: "Reviewer"}
	URLColumn            = pfdtable.ColumnAliases{Ja: "URL", En: "URL"}
)

type Config struct {
	Locale               locale.Locale
	FormatColumn         string
	ReviewCriteriaColumn string
	ReviewerColumn       string
	URLColumn            string
}

func DefaultConfig(l locale.Locale) Config {
	return Config{
		Locale:               l,
		FormatColumn:         FormatColumn.Canonical(l),
		ReviewCriteriaColumn: ReviewCriteriaColumn.Canonical(l),
		ReviewerColumn:       ReviewerColumn.Canonical(l),
		URLColumn:            URLColumn.Canonical(l),
	}
}

type Ticket struct {
	ID          string `json:"id"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

type deliverable struct {
	id             string
	description    string
	format         string
	reviewCriteria string
	reviewer       string
	url            string
	location       string
}

type messages struct {
	inputHeading     string
	outputHeading    string
	qualityLabel     string
	reviewerLabel    string
	linkLabel        string
	formatLabel      string
	placeholder      string
	none             string
	locationLabel    string
	locationProduced string
	locationUnknown  string
}

func messagesFor(l locale.Locale) messages {
	switch l {
	case locale.LocaleEn:
		return messages{
			inputHeading:     "# Input Deliverables",
			outputHeading:    "# Output Deliverables with Quality Criteria and Reviewers",
			qualityLabel:     "Quality Criteria",
			reviewerLabel:    "Reviewer",
			linkLabel:        "Deliverable Link",
			formatLabel:      "Format",
			placeholder:      PlaceholderDeliverableLinkEn,
			none:             "(none)",
			locationLabel:    "Location",
			locationProduced: LocationProducedTemplateEn,
			locationUnknown:  LocationUnknownEn,
		}
	default:
		return messages{
			inputHeading:     "# 入力成果物の一覧",
			outputHeading:    "# 出力成果物の一覧と品質基準とレビューア",
			qualityLabel:     "品質基準",
			reviewerLabel:    "レビューア",
			linkLabel:        "成果物リンク",
			formatLabel:      "フォーマット",
			placeholder:      PlaceholderDeliverableLinkJa,
			none:             "（なし）",
			locationLabel:    "所在",
			locationProduced: LocationProducedTemplateJa,
			locationUnknown:  LocationUnknownJa,
		}
	}
}

func GenerateTickets(p *pfd.PFD, ad *pfd.AtomicDeliverableTable, config Config, logger *slog.Logger) []Ticket {
	nodeMap := pfd.NewNodeMap(p.Nodes, logger)
	resolve := newDeliverableResolver(ad, config, nodeMap)
	msg := messagesFor(config.Locale)

	processIDs := make([]pfd.NodeID, 0, p.Nodes.Len())
	for _, node := range p.Nodes.Iter() {
		if node.Type != pfd.NodeTypeAtomicProcess {
			continue
		}
		if node.ID == pfd.NodeIDContextDiagram {
			continue
		}
		processIDs = append(processIDs, node.ID)
	}
	slices.SortFunc(processIDs, pfd.NodeID.Compare)

	tickets := make([]Ticket, 0, len(processIDs))
	for _, processID := range processIDs {
		node := nodeMap[processID]
		inputs := resolve(p.InputsExceptFeedback(processID))
		for i := range inputs {
			inputs[i].location = inputLocation(p, nodeMap, processID, inputs[i], msg)
		}
		tickets = append(tickets, Ticket{
			ID:          string(processID),
			Summary:     fmt.Sprintf("%s: %s", processID, node.Description),
			Description: buildDescription(inputs, resolve(p.OutputsExceptFeedback(processID)), msg),
		})
	}
	return tickets
}

func inputLocation(p *pfd.PFD, nodeMap map[pfd.NodeID]*pfd.Node, processID pfd.NodeID, d deliverable, msg messages) string {
	if producer, ok := p.ProducingAtomicProcess(pfd.NodeID(d.id), nodeMap, processID); ok {
		return fmt.Sprintf(msg.locationProduced, producer)
	}
	if d.url != "" {
		return msg.locationLabel + ": " + d.url
	}
	return msg.locationLabel + ": " + msg.locationUnknown
}

func newDeliverableResolver(ad *pfd.AtomicDeliverableTable, config Config, nodeMap map[pfd.NodeID]*pfd.Node) func(*sets.Set[pfd.NodeID]) []deliverable {
	rowByID := make(map[string]*pfd.AtomicDeliverableRow)
	idxFormat, idxReviewCriteria, idxReviewer, idxURL := -1, -1, -1, -1
	if ad != nil {
		for _, row := range ad.Rows {
			rowByID[string(row.ID)] = row
		}
		idxFormat = slices.Index(ad.ExtraHeaders, config.FormatColumn)
		idxReviewCriteria = slices.Index(ad.ExtraHeaders, config.ReviewCriteriaColumn)
		idxReviewer = slices.Index(ad.ExtraHeaders, config.ReviewerColumn)
		idxURL = slices.Index(ad.ExtraHeaders, config.URLColumn)
	}

	return func(ids *sets.Set[pfd.NodeID]) []deliverable {
		deliverables := make([]deliverable, 0, ids.Len())
		for _, id := range ids.Iter() {
			d := deliverable{id: string(id)}
			if row, ok := rowByID[string(id)]; ok {
				d.description = row.Description
				d.format = cell(row.ExtraCells, idxFormat)
				d.reviewCriteria = cell(row.ExtraCells, idxReviewCriteria)
				d.reviewer = cell(row.ExtraCells, idxReviewer)
				d.url = cell(row.ExtraCells, idxURL)
			} else if node, ok := nodeMap[id]; ok {
				d.description = node.Description
			}
			deliverables = append(deliverables, d)
		}
		return deliverables
	}
}

func cell(cells []string, idx int) string {
	if idx < 0 || idx >= len(cells) {
		return ""
	}
	return cells[idx]
}

func buildDescription(inputs []deliverable, outputs []deliverable, msg messages) string {
	sb := &strings.Builder{}

	sb.WriteString(msg.inputHeading + "\n\n")
	if len(inputs) == 0 {
		sb.WriteString(msg.none + "\n")
	} else {
		for _, d := range inputs {
			sb.WriteString(fmt.Sprintf("- %s: %s", d.id, d.description))
			if extra := inputExtra(d, msg); extra != "" {
				sb.WriteString("（" + extra + "）")
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n" + msg.outputHeading + "\n\n")
	if len(outputs) == 0 {
		sb.WriteString(msg.none + "\n")
	} else {
		for _, d := range outputs {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", d.id, d.description))
			if d.reviewCriteria != "" {
				sb.WriteString("  - " + msg.qualityLabel + ": " + d.reviewCriteria + "\n")
			}
			if d.reviewer != "" {
				sb.WriteString("  - " + msg.reviewerLabel + ": " + d.reviewer + "\n")
			}
			sb.WriteString("  - " + msg.linkLabel + ": " + deliverableLink(d, msg) + "\n")
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

func inputExtra(d deliverable, msg messages) string {
	parts := make([]string, 0, 2)
	if d.format != "" {
		parts = append(parts, msg.formatLabel+": "+d.format)
	}

	if d.location != "" {
		parts = append(parts, d.location)
	}
	return strings.Join(parts, "／")
}

func deliverableLink(d deliverable, msg messages) string {
	if d.url != "" {
		return d.url
	}
	return msg.placeholder
}

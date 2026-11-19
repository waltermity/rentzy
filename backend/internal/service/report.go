package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/shopspring/decimal"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/repository"
)

type ReportService struct {
	reports repository.ReportRepository
	now     Clock
}

func NewReportService(reports repository.ReportRepository, now Clock) *ReportService {
	return &ReportService{reports: reports, now: now}
}

// DefaultReportWindow is used when the caller gives no period.
const DefaultReportWindow = 90 * 24 * time.Hour

// Window is a reporting period [From, To).
type Window struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// ResolveWindow fills missing ends: To defaults to now, From to To - 90 days.
func (s *ReportService) ResolveWindow(from, to *time.Time) (Window, error) {
	w := Window{To: s.now()}
	if to != nil {
		w.To = *to
	}
	w.From = w.To.Add(-DefaultReportWindow)
	if from != nil {
		w.From = *from
	}
	if !w.To.After(w.From) {
		return w, invalid("to must be after from")
	}
	return w, nil
}

// TypeStats aggregates ItemStats per tool type.
type TypeStats struct {
	ToolTypeID   int64           `json:"tool_type_id"`
	ToolTypeName string          `json:"tool_type_name"`
	CategoryName string          `json:"category_name"`
	ItemCount    int             `json:"item_count"`     // items that are not retired
	RentalCount  int             `json:"rental_count"`   // rentals started in the window
	RentedHours  float64         `json:"rented_hours"`   // over non-retired items
	Utilization  float64         `json:"utilization"`    // rented / capacity hours, 0..1
	Revenue      decimal.Decimal `json:"revenue"`        // fees of rentals returned in the window
	PurchaseCost decimal.Decimal `json:"purchase_price"` // per item
	capacity     float64
}

// ItemUtilization is one item's figures with its utilization.
type ItemUtilization struct {
	domain.ItemStats
	Utilization float64 `json:"utilization"`
}

func ratio(a, b float64) float64 {
	if b <= 0 {
		return 0
	}
	return math.Round(a/b*10000) / 10000
}

func aggregate(stats []domain.ItemStats) []TypeStats {
	byType := map[int64]*TypeStats{}
	var order []int64
	for _, s := range stats {
		t, ok := byType[s.ToolTypeID]
		if !ok {
			t = &TypeStats{ToolTypeID: s.ToolTypeID, ToolTypeName: s.ToolTypeName, CategoryName: s.CategoryName,
				PurchaseCost: s.PurchasePrice, Revenue: decimal.Zero}
			byType[s.ToolTypeID] = t
			order = append(order, s.ToolTypeID)
		}
		t.RentalCount += s.RentalCount
		t.Revenue = t.Revenue.Add(s.Revenue)
		if s.Status != domain.ToolRetired {
			t.ItemCount++
			t.RentedHours += s.RentedHours
			t.capacity += s.CapacityHours
		}
	}
	out := make([]TypeStats, 0, len(order))
	for _, id := range order {
		t := byType[id]
		t.Utilization = ratio(t.RentedHours, t.capacity)
		t.RentedHours = math.Round(t.RentedHours*100) / 100
		out = append(out, *t)
	}
	return out
}

type PopularReport struct {
	Window Window      `json:"window"`
	Types  []TypeStats `json:"types"`
}

// Popular lists tool types by the number of rentals started in the window.
func (s *ReportService) Popular(ctx context.Context, w Window, limit int) (PopularReport, error) {
	stats, err := s.reports.ItemStats(ctx, w.From, w.To)
	if err != nil {
		return PopularReport{}, err
	}
	types := aggregate(stats)
	sort.SliceStable(types, func(i, j int) bool {
		if types[i].RentalCount != types[j].RentalCount {
			return types[i].RentalCount > types[j].RentalCount
		}
		return types[i].Revenue.GreaterThan(types[j].Revenue)
	})
	if limit > 0 && len(types) > limit {
		types = types[:limit]
	}
	return PopularReport{Window: w, Types: types}, nil
}

type UtilizationReport struct {
	Window       Window            `json:"window"`
	TotalRevenue decimal.Decimal   `json:"total_revenue"`
	Types        []TypeStats       `json:"types"`
	Items        []ItemUtilization `json:"items"`
}

// Utilization gives utilization and revenue per tool type and per item.
func (s *ReportService) Utilization(ctx context.Context, w Window) (UtilizationReport, error) {
	stats, err := s.reports.ItemStats(ctx, w.From, w.To)
	if err != nil {
		return UtilizationReport{}, err
	}
	rep := UtilizationReport{Window: w, TotalRevenue: decimal.Zero, Types: aggregate(stats),
		Items: make([]ItemUtilization, 0, len(stats))}
	for _, st := range stats {
		rep.TotalRevenue = rep.TotalRevenue.Add(st.Revenue)
		st.RentedHours = math.Round(st.RentedHours*100) / 100
		st.CapacityHours = math.Round(st.CapacityHours*100) / 100
		rep.Items = append(rep.Items, ItemUtilization{ItemStats: st, Utilization: ratio(st.RentedHours, st.CapacityHours)})
	}
	return rep, nil
}

// ProcurementSuggestion recommends buying more items of a tool type.
type ProcurementSuggestion struct {
	TypeStats
	SuggestedAdditional int    `json:"suggested_additional"`
	Reason              string `json:"reason"`
	// Days to earn back one item's purchase price at the window's revenue per item.
	PaybackDays *int `json:"payback_days"`
}

type ProcurementReport struct {
	Window      Window                  `json:"window"`
	Threshold   float64                 `json:"threshold"`
	Suggestions []ProcurementSuggestion `json:"suggestions"`
}

// DefaultProcurementThreshold is the utilization at which more items are suggested.
const DefaultProcurementThreshold = 0.7

// Procurement suggests buying more of every tool type whose non-retired items
// were, on average, rented for at least threshold of the window. The suggested
// number is the smallest that would bring utilization below the threshold.
func (s *ReportService) Procurement(ctx context.Context, w Window, threshold float64) (ProcurementReport, error) {
	if threshold <= 0 || threshold > 1 {
		return ProcurementReport{}, invalid("threshold must be in (0, 1]")
	}
	stats, err := s.reports.ItemStats(ctx, w.From, w.To)
	if err != nil {
		return ProcurementReport{}, err
	}
	rep := ProcurementReport{Window: w, Threshold: threshold, Suggestions: []ProcurementSuggestion{}}
	for _, t := range aggregate(stats) {
		if t.ItemCount == 0 || t.capacity <= 0 || t.Utilization < threshold {
			continue
		}
		perItemCapacity := t.capacity / float64(t.ItemCount)
		needed := int(math.Floor(t.RentedHours/(threshold*perItemCapacity))) + 1
		extra := needed - t.ItemCount
		if extra < 1 {
			extra = 1
		}
		sug := ProcurementSuggestion{TypeStats: t, SuggestedAdditional: extra,
			Reason: fmt.Sprintf("utilization %.0f%% of %d item(s) is at or above the %.0f%% threshold",
				t.Utilization*100, t.ItemCount, threshold*100)}
		windowDays := w.To.Sub(w.From).Hours() / 24
		revenuePerItemPerDay := t.Revenue.InexactFloat64() / float64(t.ItemCount) / windowDays
		if revenuePerItemPerDay > 0 {
			days := int(math.Ceil(t.PurchaseCost.InexactFloat64() / revenuePerItemPerDay))
			sug.PaybackDays = &days
		}
		rep.Suggestions = append(rep.Suggestions, sug)
	}
	sort.SliceStable(rep.Suggestions, func(i, j int) bool {
		return rep.Suggestions[i].Utilization > rep.Suggestions[j].Utilization
	})
	return rep, nil
}

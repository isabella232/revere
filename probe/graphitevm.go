package probe

import (
	"encoding/json"
	"strconv"

	"github.com/juju/errors"

	"github.com/yext/revere/datasource"
	"github.com/yext/revere/db"
	"github.com/yext/revere/util"
)

type GraphiteThresholdType struct{}

type GraphiteThresholdProbe struct {
	GraphiteThresholdType

	// TODO(fchen): fix tags on front-end js
	URL               string
	SourceID          db.DatasourceID
	Expression        string
	Thresholds        ThresholdsModel
	AuditFunction     string
	CheckPeriod       int64
	CheckPeriodType   string
	TriggerIf         string
	AuditPeriod       int64
	AuditPeriodType   string
	IgnoredPeriod     int64
	IgnoredPeriodType string
}

type ThresholdsModel struct {
	Warning  *float64
	Error    *float64
	Critical *float64
}

var (
	validGraphitePeriodTypes = []string{
		"day",
		"hour",
		"minute",
		"second",
	}
)

func init() {
	addProbeVMType(GraphiteThresholdType{})
}

func (GraphiteThresholdType) Id() db.ProbeType {
	return 1
}

func (GraphiteThresholdType) Name() string {
	return "Graphite Threshold"
}

func (GraphiteThresholdType) loadFromParams(probe string) (ProbeVM, error) {
	var g GraphiteThresholdProbe
	err := json.Unmarshal([]byte(probe), &g)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (GraphiteThresholdType) loadFromDb(encodedProbe string, tx *db.Tx) (ProbeVM, error) {
	var g GraphiteThresholdDBModel
	err := json.Unmarshal([]byte(encodedProbe), &g)
	if err != nil {
		return nil, err
	}

	checkPeriod, checkPeriodType := util.GetPeriodAndType(g.CheckPeriodMilli)
	auditPeriod, auditPeriodType := util.GetPeriodAndType(g.TimeToAuditMilli)
	ignoredPeriod, ignoredPeriodType := util.GetPeriodAndType(g.RecentTimeToIgnoreMilli)

	dbds, err := tx.LoadDatasource(db.DatasourceID(g.SourceID))
	if err != nil {
		return nil, err
	}

	if dbds == nil {
		return nil, errors.Errorf("no data source found: %d")
	}

	ds, err := datasource.LoadFromDB(datasource.GraphiteDataSource{}.Id(), dbds.Source)
	if err != nil {
		return nil, err
	}

	gds, found := ds.(*datasource.GraphiteDataSource)
	if !found {
		return nil, errors.New("not a graphite data source")
	}

	return &GraphiteThresholdProbe{
		URL:        gds.URL,
		SourceID:   db.DatasourceID(g.SourceID),
		Expression: g.Expression,
		Thresholds: ThresholdsModel{
			g.Thresholds.Warning,
			g.Thresholds.Error,
			g.Thresholds.Critical,
		},
		AuditFunction:     g.AuditFunction,
		CheckPeriod:       checkPeriod,
		CheckPeriodType:   checkPeriodType,
		TriggerIf:         g.TriggerIf,
		AuditPeriod:       auditPeriod,
		AuditPeriodType:   auditPeriodType,
		IgnoredPeriod:     ignoredPeriod,
		IgnoredPeriodType: ignoredPeriodType,
	}, nil
}

func (GraphiteThresholdType) blank() (ProbeVM, error) {
	return &GraphiteThresholdProbe{}, nil
}

func (GraphiteThresholdType) Templates() map[string]string {
	return map[string]string{
		"edit": "graphite-edit.html",
		"view": "graphite-view.html",
	}
}

func (gt GraphiteThresholdType) Scripts() map[string][]string {
	return map[string][]string{
		"edit": []string{
			"graphite-threshold.js",
			"graphite-ds-loader.js",
		},
		"preview": []string{
			"graphite-preview.js",
		},
	}
}

func (GraphiteThresholdType) AcceptedSourceTypes() []db.SourceType {
	return []db.SourceType{
		datasource.Graphite{}.Id(),
	}
}

func (g GraphiteThresholdProbe) HasDatasource(id db.DatasourceID) bool {
	return g.SourceID == id
}

func (g GraphiteThresholdProbe) SerializeForFrontend() map[string]string {
	var warningStr, errorStr, criticalStr string
	if g.Thresholds.Warning != nil {
		warningStr = strconv.FormatFloat(*g.Thresholds.Warning, 'f', -1, 64)
	}
	if g.Thresholds.Error != nil {
		errorStr = strconv.FormatFloat(*g.Thresholds.Error, 'f', -1, 64)
	}
	if g.Thresholds.Critical != nil {
		criticalStr = strconv.FormatFloat(*g.Thresholds.Critical, 'f', -1, 64)
	}
	return map[string]string{
		"Expression": g.Expression,
		"URL":        g.URL,
		"Warning":    warningStr,
		"Error":      errorStr,
		"Critical":   criticalStr,
	}
}

func (g GraphiteThresholdProbe) SerializeForDB() (string, error) {
	checkPeriodMilli := util.GetMs(g.CheckPeriod, g.CheckPeriodType)
	auditPeriodMilli := util.GetMs(g.AuditPeriod, g.AuditPeriodType)
	ignoredPeriodMilli := util.GetMs(g.IgnoredPeriod, g.IgnoredPeriodType)

	gtDB := GraphiteThresholdDBModel{
		SourceID:   int64(g.SourceID),
		Expression: g.Expression,
		Thresholds: GraphiteThresholdThresholdsDBModel{
			Warning:  g.Thresholds.Warning,
			Error:    g.Thresholds.Error,
			Critical: g.Thresholds.Critical,
		},
		TriggerIf:               g.TriggerIf,
		CheckPeriodMilli:        checkPeriodMilli,
		TimeToAuditMilli:        auditPeriodMilli,
		RecentTimeToIgnoreMilli: ignoredPeriodMilli,
		AuditFunction:           g.AuditFunction,
	}

	gtDBJSON, err := json.Marshal(gtDB)
	return string(gtDBJSON), err
}

// TODO(fchen): fix references to ProbeType() in frontend
func (g GraphiteThresholdProbe) Type() ProbeVMType {
	return GraphiteThresholdType{}
}

func (g GraphiteThresholdProbe) Validate() (errs []string) {
	if g.Expression == "" {
		errs = append(errs, "Graphite expression is required")
	}

	isValidCheckPeriodType := false
	for _, vpt := range validGraphitePeriodTypes {
		if g.CheckPeriodType == vpt {
			isValidCheckPeriodType = true
			break
		}
	}
	if !isValidCheckPeriodType {
		errs = append(errs, "Invalid check period type")
	}

	isValidAuditPeriodType := false
	for _, vpt := range validGraphitePeriodTypes {
		if g.AuditPeriodType == vpt {
			isValidAuditPeriodType = true
			break
		}
	}
	if !isValidAuditPeriodType {
		errs = append(errs, "Invalid audit period type")
	}

	if util.GetMs(g.CheckPeriod, g.CheckPeriodType) <= 0 {
		errs = append(errs, "Invalid check period")
	}

	if util.GetMs(g.AuditPeriod, g.AuditPeriodType) <= 0 {
		errs = append(errs, "Invalid audit period")
	}

	return
}
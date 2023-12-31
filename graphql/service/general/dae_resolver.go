/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package general

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/daeuniverse/dae-wing/db"
)

type DaeResolver struct {
	Ctx context.Context
}

func (r *DaeResolver) Running() (bool, error) {
	var m db.System
	q := db.DB(r.Ctx).Select("running").Model(&db.System{}).FirstOrCreate(&m)
	if q.Error != nil {
		return false, q.Error
	}
	return m.Running, nil
}

func (r *DaeResolver) Modified() (bool, error) {
	var m db.System
	tx := db.BeginReadOnlyTx(r.Ctx)
	defer tx.Commit()
	q := tx.Model(&m).
		Preload("RunningGroups").
		FirstOrCreate(&m)
	if q.Error != nil {
		return false, q.Error
	}
	if !m.Running {
		return false, nil
	}
	var selectedConfig db.Config
	if q = tx.Model(&db.Config{}).Where("selected = ?", true).First(&selectedConfig); q.Error != nil || q.RowsAffected == 0 {
		// No selected config. Maybe the running config was deleted.
		return true, q.Error
	}
	var selectedDns db.Dns
	if q = tx.Model(&db.Dns{}).Where("selected = ?", true).First(&selectedDns); q.Error != nil || q.RowsAffected == 0 {
		// No selected dns. Maybe the running dns was deleted.
		return true, q.Error
	}
	var selectedRouting db.Routing
	if q = tx.Model(&db.Routing{}).Where("selected = ?", true).First(&selectedRouting); q.Error != nil || q.RowsAffected == 0 {
		// No selected routing. Maybe the running routing was deleted.
		return true, q.Error
	}

	if selectedConfig.ID != *m.RunningConfigID || selectedConfig.Version != m.RunningConfigVersion ||
		selectedDns.ID != *m.RunningDnsID || selectedDns.Version != m.RunningDnsVersion ||
		selectedRouting.ID != *m.RunningRoutingID || selectedRouting.Version != m.RunningRoutingVersion ||
		len(m.RunningGroups) == 0 {
		return true, nil
	}
	var gvs uint
	var gids []string
	for _, g := range m.RunningGroups {
		gvs += g.Version
		gids = append(gids, fmt.Sprintf("%x", g.ID))
	}
	sort.Slice(gids, func(i, j int) bool {
		return gids[i] < gids[j]
	})
	if gvs != m.RunningGroupVersionSum || strings.Join(gids, ",") != m.RunningGroupIds {
		return true, nil
	}
	return false, nil
}

func (r *DaeResolver) Version() string {
	return db.AppVersion
}

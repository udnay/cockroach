// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package migrations contains the implementation of migrations. It is imported
// by the server library.
//
// This package registers the migrations with the migration package.
package migrations

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/errors"
)

// GetMigration returns the migration corresponding to this version if
// one exists.
func GetMigration(key clusterversion.ClusterVersion) (migration.Migration, bool) {
	m, ok := registry[key]
	return m, ok
}

// NoPrecondition is a PreconditionFunc that doesn't check anything.
func NoPrecondition(context.Context, clusterversion.ClusterVersion, migration.TenantDeps) error {
	return nil
}

// registry defines the global mapping between a cluster version and the
// associated migration. The migration is only executed after a cluster-wide
// bump of the corresponding version gate.
var registry = make(map[clusterversion.ClusterVersion]migration.Migration)

var migrations = []migration.Migration{
	migration.NewSystemMigration(
		"stop using monolithic encryption-at-rest registry for all stores",
		toCV(clusterversion.RecordsBasedRegistry),
		recordsBasedRegistryMigration,
	),
	migration.NewTenantMigration(
		"add the systems.join_tokens table",
		toCV(clusterversion.JoinTokensTable),
		NoPrecondition,
		joinTokensTableMigration,
	),
	migration.NewTenantMigration(
		"delete the deprecated namespace table descriptor at ID=2",
		toCV(clusterversion.DeleteDeprecatedNamespaceTableDescriptorMigration),
		NoPrecondition,
		deleteDeprecatedNamespaceTableDescriptorMigration,
	),
	migration.NewTenantMigration(
		"fix all descriptors",
		toCV(clusterversion.FixDescriptors),
		NoPrecondition,
		fixDescriptorMigration,
	),
	migration.NewTenantMigration(
		"add the system.database_role_settings table",
		toCV(clusterversion.DatabaseRoleSettings),
		NoPrecondition,
		databaseRoleSettingsTableMigration,
	),
	migration.NewTenantMigration(
		"add the system.tenant_usage table",
		toCV(clusterversion.TenantUsageTable),
		NoPrecondition,
		tenantUsageTableMigration,
	),
	migration.NewTenantMigration(
		"add the system.sql_instances table",
		toCV(clusterversion.SQLInstancesTable),
		NoPrecondition,
		sqlInstancesTableMigration,
	),
	migration.NewSystemMigration(
		"move over all intents to separate lock table",
		toCV(clusterversion.SeparatedIntentsMigration),
		separatedIntentsMigration),
	migration.NewSystemMigration(
		"run no-op migrate command on all ranges after lock table migration",
		toCV(clusterversion.PostSeparatedIntentsMigration),
		postSeparatedIntentsMigration),
	migration.NewTenantMigration(
		"add last_run and num_runs columns to system.jobs",
		toCV(clusterversion.RetryJobsWithExponentialBackoff),
		NoPrecondition,
		retryJobsWithExponentialBackoff),
	migration.NewTenantMigration(
		"add system.zones table for secondary tenants",
		toCV(clusterversion.ZonesTableForSecondaryTenants),
		NoPrecondition,
		zonesTableForSecondaryTenants,
	),
	migration.NewTenantMigration(
		"add the system.span_configurations table to system tenant",
		toCV(clusterversion.SpanConfigurationsTable),
		NoPrecondition,
		spanConfigurationsTableMigration,
	),
	migration.NewTenantMigration(
		"create indexes on revokedAt and lastUsedAt columns from system.web_sessions",
		toCV(clusterversion.AlterSystemWebSessionsCreateIndexes),
		NoPrecondition,
		alterSystemWebSessionsCreateIndexes,
	),
	migration.NewTenantMigration(
		"change system.tenant_usage table to use a single column for consumption",
		toCV(clusterversion.TenantUsageSingleConsumptionColumn),
		NoPrecondition,
		tenantUsageSingleConsumptionColumn,
	),
	migration.NewTenantMigration(
		"add the system.statement_statistics and system.transaction_statistics tables",
		toCV(clusterversion.SQLStatsTables),
		NoPrecondition,
		sqlStatsTablesMigration,
	),
	migration.NewTenantMigration(
		"ensure that draining names are no longer in use",
		toCV(clusterversion.DrainingNamesMigration),
		NoPrecondition,
		ensureNoDrainingNames,
	),
	migration.NewTenantMigration(
		"add column avgSize to table system.table_statistics",
		toCV(clusterversion.AlterSystemTableStatisticsAddAvgSizeCol),
		NoPrecondition,
		alterSystemTableStatisticsAddAvgSize,
	),
	migration.NewTenantMigration(
		"update system.statement_diagnostics_requests table to support conditional stmt diagnostics",
		toCV(clusterversion.AlterSystemStmtDiagReqs),
		NoPrecondition,
		alterSystemStmtDiagReqs,
	),
	migration.NewTenantMigration(
		"seed system.span_configurations with configs for existing for existing tenants",
		toCV(clusterversion.SeedTenantSpanConfigs),
		NoPrecondition,
		seedTenantSpanConfigsMigration,
	),
	migration.NewTenantMigration("insert missing system.namespace entries for public schemas",
		toCV(clusterversion.InsertPublicSchemaNamespaceEntryOnRestore),
		NoPrecondition,
		insertMissingPublicSchemaNamespaceEntry,
	),
}

func init() {
	for _, m := range migrations {
		if _, exists := registry[m.ClusterVersion()]; exists {
			panic(errors.AssertionFailedf("duplicate migration registration for %v", m.ClusterVersion()))
		}
		registry[m.ClusterVersion()] = m
	}
}

func toCV(key clusterversion.Key) clusterversion.ClusterVersion {
	return clusterversion.ClusterVersion{
		Version: clusterversion.ByKey(key),
	}
}

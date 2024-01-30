// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

//go:build !plan9

package kube

import (
	"slices"

	"go.uber.org/zap"
	xslices "golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	tsapi "tailscale.com/k8s-operator/apis/v1alpha1"
	"tailscale.com/tstime"
)

// SetConnectorCondition ensures that Connector status has a condition with the
// given attributes. LastTransitionTime gets set every time condition's status
// changes
func SetConnectorCondition(cn *tsapi.Connector, conditionType tsapi.ConnectorConditionType, status metav1.ConditionStatus, reason, message string, gen int64, clock tstime.Clock, logger *zap.SugaredLogger) {
	conds := updateCondition(cn.Status.Conditions, conditionType, status, reason, message, gen, clock, logger)
	cn.Status.Conditions = conds
}

// RemoveConnectorCondition will remove condition of the given type
func RemoveConnectorCondition(conn *tsapi.Connector, conditionType tsapi.ConnectorConditionType) {
	conn.Status.Conditions = slices.DeleteFunc(conn.Status.Conditions, func(cond tsapi.ConnectorCondition) bool {
		return cond.Type == conditionType
	})
}

// SetDNSConfigCondition ensures that DNSConfig status has a condition with the
// given attributes. LastTransitionTime gets set every time condition's status
// changes
func SetDNSConfigCondition(dnsCfg *tsapi.DNSConfig, conditionType tsapi.ConnectorConditionType, status metav1.ConditionStatus, reason, message string, gen int64, clock tstime.Clock, logger *zap.SugaredLogger) {
	conds := updateCondition(dnsCfg.Status.Conditions, conditionType, status, reason, message, gen, clock, logger)
	dnsCfg.Status.Conditions = conds
}

func updateCondition(conds []tsapi.ConnectorCondition, conditionType tsapi.ConnectorConditionType, status metav1.ConditionStatus, reason, message string, gen int64, clock tstime.Clock, logger *zap.SugaredLogger) []tsapi.ConnectorCondition {
	newCondition := tsapi.ConnectorCondition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: gen,
	}

	nowTime := metav1.NewTime(clock.Now())
	newCondition.LastTransitionTime = &nowTime

	idx := xslices.IndexFunc(conds, func(cond tsapi.ConnectorCondition) bool {
		return cond.Type == conditionType
	})

	if idx == -1 {
		conds = append(conds, newCondition)
		return conds
	}

	// Update the existing condition
	cond := conds[idx]
	// If this update doesn't contain a state transition, we don't update
	// the conditions LastTransitionTime to Now()
	if cond.Status == status {
		newCondition.LastTransitionTime = cond.LastTransitionTime
	} else {
		logger.Info("Status change for condition %s from %s to %s", conditionType, cond.Status, status)
	}
	return conds
}

func DNSCfgIsReady(cfg *tsapi.DNSConfig) bool {
	idx := xslices.IndexFunc(cfg.Status.Conditions, func(cond tsapi.ConnectorCondition) bool {
		return cond.Type == tsapi.NameserverReady
	})
	if idx == -1 {
		return false
	}
	cond := cfg.Status.Conditions[idx]
	return cond.Status == metav1.ConditionTrue && cond.ObservedGeneration == cfg.Generation
}

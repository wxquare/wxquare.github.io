package registry

import (
	"context"
	"sort"

	"common-services/internal/idgen"
)

type StaticRegistry struct {
	byNamespace map[string]idgen.NamespaceConfig
}

func NewStatic(configs []idgen.NamespaceConfig) *StaticRegistry {
	byNamespace := make(map[string]idgen.NamespaceConfig, len(configs))
	for _, cfg := range configs {
		byNamespace[cfg.Namespace] = cfg
	}
	return &StaticRegistry{byNamespace: byNamespace}
}

func (r *StaticRegistry) Get(ctx context.Context, namespace string) (idgen.NamespaceConfig, error) {
	cfg, ok := r.byNamespace[namespace]
	if !ok {
		return idgen.NamespaceConfig{}, idgen.NewError(idgen.ErrNamespaceNotFound, namespace, "namespace is not registered", false)
	}
	if cfg.Status != idgen.NamespaceEnabled {
		return idgen.NamespaceConfig{}, idgen.NewError(idgen.ErrNamespaceDisabled, namespace, "namespace is not enabled", false)
	}
	return cfg, nil
}

func (r *StaticRegistry) List(ctx context.Context) ([]idgen.NamespaceConfig, error) {
	result := make([]idgen.NamespaceConfig, 0, len(r.byNamespace))
	for _, cfg := range r.byNamespace {
		result = append(result, cfg)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Namespace < result[j].Namespace
	})
	return result, nil
}

func DefaultNamespaces() []idgen.NamespaceConfig {
	return []idgen.NamespaceConfig{
		segment("product.item", "product", "item", 1000, 800000000000),
		segment("product.spu", "product", "spu", 1000, 700000000000),
		segment("product.sku", "product", "sku", 1000, 600000000000),
		ulid("supply.draft", "supply", "draft"),
		ulid("supply.staging", "supply", "staging"),
		ulid("supply.qc_review", "supply", "qc"),
		ulid("checkout.session", "trade", "chk"),
		business("trade.order", "trade", "ORD"),
		business("trade.payment", "trade", "PAY"),
		business("trade.refund", "trade", "RF"),
		snowflake("inventory.ledger", "inventory", "ledger"),
		ulid("inventory.reservation", "inventory", "rsrv"),
		ulid("event.outbox", "event", "evt"),
		ulid("trace.operation", "trace", "op"),
	}
}

func segment(ns, domain, prefix string, step, start int64) idgen.NamespaceConfig {
	return idgen.NamespaceConfig{
		Namespace: ns, BizDomain: domain, IDType: idgen.IDTypeInt64,
		GeneratorType: idgen.GeneratorSegment, Prefix: prefix,
		ExposeScope: idgen.ExposeInternal, Step: step, MaxCapacity: start,
		OwnerTeam: domain + "-team", Status: idgen.NamespaceEnabled,
	}
}

func snowflake(ns, domain, prefix string) idgen.NamespaceConfig {
	return idgen.NamespaceConfig{
		Namespace: ns, BizDomain: domain, IDType: idgen.IDTypeInt64,
		GeneratorType: idgen.GeneratorSnowflake, Prefix: prefix,
		ExposeScope: idgen.ExposeInternal, Step: 0,
		OwnerTeam: domain + "-team", Status: idgen.NamespaceEnabled,
	}
}

func business(ns, domain, prefix string) idgen.NamespaceConfig {
	return idgen.NamespaceConfig{
		Namespace: ns, BizDomain: domain, IDType: idgen.IDTypeBusinessNo,
		GeneratorType: idgen.GeneratorSnowflake, Prefix: prefix,
		ExposeScope: idgen.ExposeMixed, Step: 0,
		OwnerTeam: domain + "-team", Status: idgen.NamespaceEnabled,
	}
}

func ulid(ns, domain, prefix string) idgen.NamespaceConfig {
	return idgen.NamespaceConfig{
		Namespace: ns, BizDomain: domain, IDType: idgen.IDTypeString,
		GeneratorType: idgen.GeneratorULID, Prefix: prefix,
		ExposeScope: idgen.ExposeInternal, Step: 0,
		OwnerTeam: domain + "-team", Status: idgen.NamespaceEnabled,
	}
}

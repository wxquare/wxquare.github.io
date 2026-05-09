package registry

import (
	"context"
	"testing"

	"common-services/internal/idgen"
)

func TestStaticRegistryReturnsDefaultNamespace(t *testing.T) {
	reg := NewStatic(DefaultNamespaces())
	cfg, err := reg.Get(context.Background(), "trade.order")
	if err != nil {
		t.Fatalf("get namespace: %v", err)
	}
	if cfg.GeneratorType != idgen.GeneratorSnowflake {
		t.Fatalf("generator = %s", cfg.GeneratorType)
	}
	if cfg.IDType != idgen.IDTypeBusinessNo {
		t.Fatalf("id type = %s", cfg.IDType)
	}
	if cfg.Prefix != "ORD" {
		t.Fatalf("prefix = %s", cfg.Prefix)
	}
}

func TestStaticRegistryRejectsMissingNamespace(t *testing.T) {
	reg := NewStatic(DefaultNamespaces())
	_, err := reg.Get(context.Background(), "unknown.namespace")
	svcErr, ok := idgen.AsServiceError(err)
	if !ok {
		t.Fatalf("expected ServiceError, got %T", err)
	}
	if svcErr.Code != idgen.ErrNamespaceNotFound {
		t.Fatalf("code = %s", svcErr.Code)
	}
}

func TestStaticRegistryRejectsDisabledNamespace(t *testing.T) {
	cfgs := DefaultNamespaces()
	cfgs[0].Status = idgen.NamespaceDisabled
	reg := NewStatic(cfgs)
	_, err := reg.Get(context.Background(), cfgs[0].Namespace)
	svcErr, ok := idgen.AsServiceError(err)
	if !ok {
		t.Fatalf("expected ServiceError, got %T", err)
	}
	if svcErr.Code != idgen.ErrNamespaceDisabled {
		t.Fatalf("code = %s", svcErr.Code)
	}
}

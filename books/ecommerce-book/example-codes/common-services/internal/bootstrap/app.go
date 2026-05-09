package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"common-services/internal/idgen/formatter"
	"common-services/internal/idgen/router"
	"common-services/internal/idgen/segment"
	"common-services/internal/idgen/snowflake"
	ulidgen "common-services/internal/idgen/ulid"
	"common-services/internal/infrastructure/lease"
	"common-services/internal/infrastructure/memory"
	"common-services/internal/infrastructure/metrics"
	mysqlstore "common-services/internal/infrastructure/mysql"
	httpapi "common-services/internal/interfaces/http"
)

type App struct {
	Addr   string
	Server *http.Server
	DB     *sql.DB
	lease  *lease.Manager
}

func NewApp(ctx context.Context) (*App, error) {
	addr := getenv("ID_SERVICE_ADDR", ":8090")
	maxBatch := getenvInt("ID_MAX_BATCH_SIZE", 1000)
	regionID := int64(getenvInt("ID_REGION_ID", 1))
	recorder := metrics.NewRecorder()

	if dsn := os.Getenv("ID_MYSQL_DSN"); dsn != "" {
		opened, err := mysqlstore.Open(ctx, dsn)
		if err != nil {
			return nil, fmt.Errorf("open mysql: %w", err)
		}
		sqlStore := mysqlstore.NewStore(opened)
		if err := sqlStore.InitSchema(ctx); err != nil {
			_ = opened.Close()
			return nil, fmt.Errorf("init mysql schema: %w", err)
		}
		instanceID := getenv("ID_INSTANCE_ID", defaultInstanceID())
		leaseManager := lease.NewManager(sqlStore, regionID, getenv("ID_DATACENTER_CODE", "local-a"), instanceID, getenvDuration("ID_WORKER_LEASE_TTL_SECONDS", 30*time.Second), getenvDuration("ID_WORKER_HEARTBEAT_SECONDS", 10*time.Second))
		if err := leaseManager.Start(ctx); err != nil {
			_ = opened.Close()
			return nil, fmt.Errorf("start worker lease: %w", err)
		}
		seg := segment.NewGenerator(sqlStore)
		sf := snowflake.NewGenerator(snowflake.Config{Epoch: defaultEpoch()}, snowflake.RealClock{}, leaseManager)
		svc := router.New(sqlStore, seg, sf, ulidgen.NewGenerator(), formatter.NewBusinessNumberFormatter(), maxBatch)
		handler := httpapi.NewHandler(svc, sqlStore, recorder, leaseManager.Ready)
		server := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
		return &App{Addr: addr, Server: server, DB: opened, lease: leaseManager}, nil
	}

	store := memory.NewStore()
	seg := segment.NewGenerator(store.Segment)
	sfLease := &snowflake.StaticLease{ReadyValue: true, WorkerIDValue: 1, RegionIDValue: regionID}
	sf := snowflake.NewGenerator(snowflake.Config{Epoch: defaultEpoch()}, snowflake.RealClock{}, sfLease)
	svc := router.New(store, seg, sf, ulidgen.NewGenerator(), formatter.NewBusinessNumberFormatter(), maxBatch)
	handler := httpapi.NewHandler(svc, store, recorder, func() bool { return true })
	server := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	return &App{Addr: addr, Server: server}, nil
}

func (a *App) Close(ctx context.Context) {
	if a.lease != nil {
		a.lease.Stop(ctx)
	}
	if a.DB != nil {
		_ = a.DB.Close()
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}

func defaultEpoch() time.Time {
	t, err := time.Parse(time.RFC3339, getenv("ID_SNOWFLAKE_EPOCH", "2026-01-01T00:00:00Z"))
	if err != nil {
		panic(fmt.Sprintf("invalid ID_SNOWFLAKE_EPOCH: %v", err))
	}
	return t
}

func defaultInstanceID() string {
	host, err := os.Hostname()
	if err != nil {
		host = "unknown-host"
	}
	return fmt.Sprintf("%s-%d-%d", host, os.Getpid(), time.Now().UnixNano())
}

package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"common-services/internal/idgen"
	"common-services/internal/idgen/registry"
	"common-services/internal/idgen/segment"
	"common-services/internal/infrastructure/audit"
	"common-services/internal/infrastructure/lease"
)

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) InitSchema(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS id_namespace (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			namespace VARCHAR(64) NOT NULL,
			biz_domain VARCHAR(64) NOT NULL,
			id_type VARCHAR(32) NOT NULL,
			generator_type VARCHAR(32) NOT NULL,
			prefix VARCHAR(32) DEFAULT NULL,
			expose_scope VARCHAR(32) NOT NULL,
			step BIGINT NOT NULL DEFAULT 1000,
			max_capacity BIGINT DEFAULT 0,
			owner_team VARCHAR(64) NOT NULL,
			status VARCHAR(32) NOT NULL,
			created_at DATETIME(6) NOT NULL,
			updated_at DATETIME(6) NOT NULL,
			UNIQUE KEY uk_namespace (namespace),
			KEY idx_domain_status (biz_domain, status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS id_segment (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			namespace VARCHAR(64) NOT NULL,
			max_id BIGINT NOT NULL,
			step BIGINT NOT NULL,
			version BIGINT NOT NULL,
			updated_at DATETIME(6) NOT NULL,
			UNIQUE KEY uk_namespace (namespace)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS id_worker (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			worker_id INT NOT NULL,
			region_id INT NOT NULL,
			datacenter_code VARCHAR(32) NOT NULL,
			instance_id VARCHAR(128) NOT NULL,
			lease_token VARCHAR(64) NOT NULL,
			lease_until DATETIME(6) NOT NULL,
			heartbeat_at DATETIME(6) NOT NULL,
			status VARCHAR(32) NOT NULL,
			created_at DATETIME(6) NOT NULL,
			updated_at DATETIME(6) NOT NULL,
			UNIQUE KEY uk_worker_region (worker_id, region_id),
			UNIQUE KEY uk_instance (instance_id),
			KEY idx_status_lease (status, lease_until)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS id_issue_log (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			request_id VARCHAR(64) NOT NULL,
			namespace VARCHAR(64) NOT NULL,
			caller VARCHAR(128) NOT NULL,
			issue_type VARCHAR(32) NOT NULL,
			issued_value VARCHAR(128) DEFAULT NULL,
			error_code VARCHAR(64) DEFAULT NULL,
			error_message VARCHAR(512) DEFAULT NULL,
			created_at DATETIME(6) NOT NULL,
			UNIQUE KEY uk_request_id (request_id),
			KEY idx_namespace_time (namespace, created_at),
			KEY idx_issue_type_time (issue_type, created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, stmt := range statements {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return s.SeedDefaults(ctx)
}

func (s *Store) SeedDefaults(ctx context.Context) error {
	now := time.Now()
	for _, cfg := range registry.DefaultNamespaces() {
		_, err := s.db.ExecContext(ctx, `INSERT INTO id_namespace(namespace,biz_domain,id_type,generator_type,prefix,expose_scope,step,max_capacity,owner_team,status,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?)
			ON DUPLICATE KEY UPDATE updated_at=VALUES(updated_at)`,
			cfg.Namespace, cfg.BizDomain, cfg.IDType, cfg.GeneratorType, nullablePrefix(cfg.Prefix), cfg.ExposeScope, cfg.Step, cfg.MaxCapacity, cfg.OwnerTeam, cfg.Status, now, now)
		if err != nil {
			return err
		}
		if cfg.GeneratorType == idgen.GeneratorSegment {
			_, err = s.db.ExecContext(ctx, `INSERT INTO id_segment(namespace,max_id,step,version,updated_at)
				VALUES(?,?,?,?,?)
				ON DUPLICATE KEY UPDATE updated_at=VALUES(updated_at)`,
				cfg.Namespace, cfg.MaxCapacity, cfg.Step, 0, now)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) Get(ctx context.Context, namespace string) (idgen.NamespaceConfig, error) {
	var cfg idgen.NamespaceConfig
	var prefix sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT namespace,biz_domain,id_type,generator_type,prefix,expose_scope,step,max_capacity,owner_team,status
		FROM id_namespace WHERE namespace=?`, namespace).Scan(
		&cfg.Namespace, &cfg.BizDomain, &cfg.IDType, &cfg.GeneratorType, &prefix,
		&cfg.ExposeScope, &cfg.Step, &cfg.MaxCapacity, &cfg.OwnerTeam, &cfg.Status,
	)
	if err == sql.ErrNoRows {
		return idgen.NamespaceConfig{}, idgen.NewError(idgen.ErrNamespaceNotFound, namespace, "namespace is not registered", false)
	}
	if err != nil {
		return idgen.NamespaceConfig{}, err
	}
	cfg.Prefix = prefix.String
	if cfg.Status != idgen.NamespaceEnabled {
		return idgen.NamespaceConfig{}, idgen.NewError(idgen.ErrNamespaceDisabled, namespace, "namespace is not enabled", false)
	}
	return cfg, nil
}

func (s *Store) List(ctx context.Context) ([]idgen.NamespaceConfig, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT namespace,biz_domain,id_type,generator_type,prefix,expose_scope,step,max_capacity,owner_team,status
		FROM id_namespace ORDER BY namespace`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]idgen.NamespaceConfig, 0)
	for rows.Next() {
		var cfg idgen.NamespaceConfig
		var prefix sql.NullString
		if err := rows.Scan(&cfg.Namespace, &cfg.BizDomain, &cfg.IDType, &cfg.GeneratorType, &prefix, &cfg.ExposeScope, &cfg.Step, &cfg.MaxCapacity, &cfg.OwnerTeam, &cfg.Status); err != nil {
			return nil, err
		}
		cfg.Prefix = prefix.String
		result = append(result, cfg)
	}
	return result, rows.Err()
}

func (s *Store) Allocate(ctx context.Context, namespace string, step int64) (segment.Range, error) {
	for attempt := 0; attempt < 3; attempt++ {
		var maxID, version int64
		if err := s.db.QueryRowContext(ctx, `SELECT max_id, version FROM id_segment WHERE namespace=?`, namespace).Scan(&maxID, &version); err != nil {
			return segment.Range{}, err
		}
		result, err := s.db.ExecContext(ctx, `UPDATE id_segment SET max_id=max_id+?, version=version+1, updated_at=? WHERE namespace=? AND version=?`, step, time.Now(), namespace, version)
		if err != nil {
			return segment.Range{}, err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return segment.Range{}, err
		}
		if affected == 1 {
			return segment.Range{Start: maxID + 1, End: maxID + step}, nil
		}
	}
	return segment.Range{}, fmt.Errorf("allocate segment conflict: namespace=%s", namespace)
}

func (s *Store) SaveIssueLog(ctx context.Context, entry audit.IssueLog) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO id_issue_log(request_id,namespace,caller,issue_type,issued_value,error_code,error_message,created_at)
		VALUES(?,?,?,?,?,?,?,?)`,
		entry.RequestID, entry.Namespace, entry.Caller, entry.IssueType, entry.IssuedValue, entry.ErrorCode, entry.ErrorMessage, entry.CreatedAt)
	return err
}

func (s *Store) AcquireWorker(ctx context.Context, regionID int64, datacenterCode, instanceID string, ttl time.Duration) (int64, string, error) {
	now := time.Now()
	until := now.Add(ttl)
	for workerID := int64(0); workerID <= 31; workerID++ {
		token := lease.NewLeaseToken()
		result, err := s.db.ExecContext(ctx, `UPDATE id_worker
			SET instance_id=?, lease_token=?, lease_until=?, heartbeat_at=?, status='ACTIVE', updated_at=?
			WHERE region_id=? AND worker_id=? AND (lease_until<? OR status<>'ACTIVE' OR instance_id=?)`,
			instanceID, token, until, now, now, regionID, workerID, now, instanceID)
		if err != nil {
			return 0, "", err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return 0, "", err
		}
		if affected == 1 {
			return workerID, token, nil
		}

		result, err = s.db.ExecContext(ctx, `INSERT IGNORE INTO id_worker(worker_id,region_id,datacenter_code,instance_id,lease_token,lease_until,heartbeat_at,status,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?,?,?,?)`,
			workerID, regionID, datacenterCode, instanceID, token, until, now, "ACTIVE", now, now)
		if err != nil {
			return 0, "", err
		}
		affected, err = result.RowsAffected()
		if err != nil {
			return 0, "", err
		}
		if affected == 1 {
			return workerID, token, nil
		}
	}
	return 0, "", fmt.Errorf("no worker lease available for region %d", regionID)
}

func (s *Store) RenewWorker(ctx context.Context, regionID, workerID int64, instanceID, leaseToken string, ttl time.Duration) error {
	now := time.Now()
	result, err := s.db.ExecContext(ctx, `UPDATE id_worker
		SET lease_until=?, heartbeat_at=?, updated_at=?
		WHERE region_id=? AND worker_id=? AND instance_id=? AND lease_token=? AND status='ACTIVE' AND lease_until>?`,
		now.Add(ttl), now, now, regionID, workerID, instanceID, leaseToken, now)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("worker lease lost: region=%d worker=%d", regionID, workerID)
	}
	return nil
}

func (s *Store) ReleaseWorker(ctx context.Context, regionID, workerID int64, instanceID, leaseToken string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE id_worker
		SET status='EXPIRED', updated_at=?
		WHERE region_id=? AND worker_id=? AND instance_id=? AND lease_token=?`,
		time.Now(), regionID, workerID, instanceID, leaseToken)
	return err
}

func nullablePrefix(prefix string) any {
	if prefix == "" {
		return nil
	}
	return prefix
}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/application/projectors"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedcontext "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/events"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	_ "github.com/lib/pq"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("recompute capability inheritance failed: %v", err)
	}
}

func run() error {
	connStr := getEnv("DB_CONN_STRING", "")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return err
	}

	tenantDB := database.NewTenantAwareDB(db)
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	eventBus := events.NewInMemoryEventBus()
	eventStore.SetEventBus(eventBus)

	capabilityRM := readmodels.NewCapabilityReadModel(tenantDB)
	realizationRM := readmodels.NewRealizationReadModel(tenantDB)
	componentCacheRM := readmodels.NewComponentCacheReadModel(tenantDB)

	realizationProjector := projectors.NewRealizationProjector(realizationRM, componentCacheRM)
	eventBus.Subscribe("CapabilityRealizationsInherited", realizationProjector)
	eventBus.Subscribe("CapabilityRealizationsUninherited", realizationProjector)

	capabilityRepo := repositories.NewCapabilityRepository(eventStore)
	handler := handlers.NewRecomputeCapabilityInheritanceHandler(capabilityRepo, capabilityRM, realizationRM)

	tenants, err := loadTenants(db)
	if err != nil {
		return err
	}

	processed := 0
	for _, tenant := range tenants {
		count, err := processTenant(tenant, capabilityRM, handler)
		if err != nil {
			return err
		}
		processed += count
	}

	log.Printf("recompute capability inheritance completed for %d capabilities", processed)
	return nil
}

func processTenant(tenant string, capabilityRM *readmodels.CapabilityReadModel, handler *handlers.RecomputeCapabilityInheritanceHandler) (int, error) {
	ctx, err := tenantContext(tenant)
	if err != nil {
		return 0, err
	}

	caps, err := capabilityRM.GetAll(ctx)
	if err != nil {
		return 0, fmt.Errorf("load capabilities for tenant %s: %w", tenant, err)
	}

	for _, cap := range caps {
		_, err := handler.Handle(ctx, &commands.RecomputeCapabilityInheritance{CapabilityID: cap.ID})
		if err != nil {
			return 0, fmt.Errorf("recompute capability %s tenant %s: %w", cap.ID, tenant, err)
		}
	}

	log.Printf("tenant %s processed %d capabilities", tenant, len(caps))
	return len(caps), nil
}

func loadTenants(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT DISTINCT tenant_id FROM capabilities ORDER BY tenant_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tenants := []string{}
	for rows.Next() {
		var tenant string
		if err := rows.Scan(&tenant); err != nil {
			return nil, err
		}
		tenants = append(tenants, tenant)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(tenants) == 0 {
		tenants = append(tenants, sharedvo.DefaultTenantID().Value())
	}
	return tenants, nil
}

func tenantContext(tenant string) (context.Context, error) {
	tenantID, err := sharedvo.NewTenantID(tenant)
	if err != nil {
		return nil, err
	}
	ctx := sharedcontext.WithTenant(context.Background(), tenantID)
	ctx = sharedcontext.WithActor(ctx, sharedcontext.Actor{ID: "system", Email: "system@easi.app"})
	return ctx, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

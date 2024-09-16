package storage

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"tender_service/internal/config"
)

type Storage struct {
	db *gorm.DB
}

func New(cancel context.CancelFunc, s *Storage, cfg *config.Config) error {
	const op = "storage.postgres.New"

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", cfg.DB.Host, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer cancel()
	db.Logger = logger.Default.LogMode(logger.Info)

	//db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

	Exec(db)

	s.db = db
	return nil
}

func Exec(db *gorm.DB) {

	db.Exec(`
			CREATE TABLE IF NOT EXISTS employee
			(
				id uuid NOT NULL DEFAULT uuid_generate_v4(),
				username character varying(50) COLLATE pg_catalog."default" NOT NULL,
				first_name character varying(50) COLLATE pg_catalog."default",
				last_name character varying(50) COLLATE pg_catalog."default",
				created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
				updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
				deleted_at timestamp with time zone,
				CONSTRAINT employee_pkey PRIMARY KEY (id),
				CONSTRAINT employee_username_key UNIQUE (username)
			);
`)

	db.Exec(`
			CREATE TABLE IF NOT EXISTS organization
			(
				id uuid NOT NULL DEFAULT uuid_generate_v4(),
				name character varying(100) COLLATE pg_catalog."default" NOT NULL,
				description text COLLATE pg_catalog."default",
				type organization_type,
				created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
				updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
				deleted_at timestamp with time zone,
				CONSTRAINT organization_pkey PRIMARY KEY (id)
			);

			CREATE INDEX IF NOT EXISTS idx_organization_deleted_at
				ON organization USING btree
				(deleted_at ASC NULLS LAST)
`)

	db.Exec(`
			CREATE TABLE IF NOT EXISTS organization_responsible
			(
				id uuid NOT NULL DEFAULT uuid_generate_v4(),
				organization_id uuid,
				user_id uuid,
				CONSTRAINT organization_responsible_pkey PRIMARY KEY (id),
				CONSTRAINT organization_responsible_organization_id_fkey FOREIGN KEY (organization_id)
					REFERENCES organization (id) MATCH SIMPLE
					ON UPDATE NO ACTION
					ON DELETE CASCADE,
				CONSTRAINT organization_responsible_user_id_fkey FOREIGN KEY (user_id)
					REFERENCES employee (id) MATCH SIMPLE
					ON UPDATE NO ACTION
					ON DELETE CASCADE
			);
`)

	db.Exec(`
			CREATE TABLE IF NOT EXISTS tender_versions
			(
				id uuid NOT NULL DEFAULT uuid_generate_v4(),
				created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
				updated_at timestamp with time zone,
				deleted_at timestamp with time zone,
				tender_id uuid,
				name character varying(100) COLLATE pg_catalog."default" NOT NULL,
				description character varying(500) COLLATE pg_catalog."default",
				service_type tender_service_type,
				status tender_status NOT NULL,
				employee_username text COLLATE pg_catalog."default" NOT NULL,
				organization_id uuid NOT NULL,
				version bigint NOT NULL DEFAULT 1,
				CONSTRAINT tender_versions_pkey PRIMARY KEY (id),
				CONSTRAINT fk_tender_versions_organization FOREIGN KEY (organization_id)
					REFERENCES organization (id) MATCH SIMPLE
					ON UPDATE NO ACTION
					ON DELETE NO ACTION
			);

			CREATE INDEX IF NOT EXISTS idx_tender_versions_deleted_at
				ON tender_versions USING btree
				(deleted_at ASC NULLS LAST);
`)

	db.Exec(`
			CREATE TABLE IF NOT EXISTS tenders
			(
				id uuid NOT NULL DEFAULT uuid_generate_v4(),
				created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
				updated_at timestamp with time zone,
				deleted_at timestamp with time zone,
				name character varying(100) COLLATE pg_catalog."default" NOT NULL,
				description character varying(500) COLLATE pg_catalog."default",
				service_type tender_service_type,
				status tender_status NOT NULL,
				employee_username text COLLATE pg_catalog."default" NOT NULL,
				organization_id uuid NOT NULL,
				version bigint NOT NULL DEFAULT 1,
				CONSTRAINT tenders_pkey PRIMARY KEY (id),
				CONSTRAINT fk_tenders_organization FOREIGN KEY (organization_id)
					REFERENCES organization (id) MATCH SIMPLE
					ON UPDATE NO ACTION
					ON DELETE NO ACTION
			);

			CREATE INDEX IF NOT EXISTS idx_tenders_deleted_at
				ON tenders USING btree
				(deleted_at ASC NULLS LAST);
`)

	db.Exec(`
			CREATE TABLE IF NOT EXISTS bid_feedbacks
			(
				id uuid NOT NULL DEFAULT uuid_generate_v4(),
				created_at timestamp with time zone,
				updated_at timestamp with time zone,
				deleted_at timestamp with time zone,
				feedback character varying(1000) COLLATE pg_catalog."default" NOT NULL,
				bid_id uuid NOT NULL,
				employee_username text COLLATE pg_catalog."default" NOT NULL,
				organization_id uuid NOT NULL,
				CONSTRAINT bid_feedbacks_pkey PRIMARY KEY (id),
				CONSTRAINT fk_bid_feedbacks_bid FOREIGN KEY (bid_id)
					REFERENCES bids (id) MATCH SIMPLE
					ON UPDATE NO ACTION
					ON DELETE NO ACTION,
				CONSTRAINT fk_bid_feedbacks_organization FOREIGN KEY (organization_id)
					REFERENCES organization (id) MATCH SIMPLE
					ON UPDATE NO ACTION
					ON DELETE NO ACTION
			);

			CREATE INDEX IF NOT EXISTS idx_bid_feedbacks_deleted_at
			    ON bid_feedbacks USING btree
				(deleted_at ASC NULLS LAST)

`)

	db.Exec(`
			CREATE TABLE IF NOT EXISTS bid_versions
			(
				id uuid NOT NULL DEFAULT uuid_generate_v4(),
				created_at timestamp with time zone,
				updated_at timestamp with time zone,
				deleted_at timestamp with time zone,
				name character varying(100) COLLATE pg_catalog."default" NOT NULL,
				description character varying(500) COLLATE pg_catalog."default",
				status bid_status NOT NULL,
				tender_id uuid,
				bid_id uuid,
				employee_username text COLLATE pg_catalog."default" NOT NULL,
				organization_id uuid NOT NULL,
				author_type bid_author_type NOT NULL,
				version bigint DEFAULT 1,
				CONSTRAINT bid_versions_pkey PRIMARY KEY (id),
				CONSTRAINT fk_bid_versions_bid FOREIGN KEY (bid_id)
					REFERENCES bids (id) MATCH SIMPLE
					ON UPDATE NO ACTION
					ON DELETE NO ACTION,
				CONSTRAINT fk_bid_versions_organization FOREIGN KEY (organization_id)
					REFERENCES organization (id) MATCH SIMPLE
					ON UPDATE NO ACTION
					ON DELETE NO ACTION,
				CONSTRAINT fk_bid_versions_tender FOREIGN KEY (tender_id)
					REFERENCES tenders (id) MATCH SIMPLE
					ON UPDATE NO ACTION
					ON DELETE NO ACTION
			);

			CREATE INDEX IF NOT EXISTS idx_bid_versions_deleted_at
				ON bid_versions USING btree
				(deleted_at ASC NULLS LAST)
`)

	db.Exec(`
			CREATE TABLE IF NOT EXISTS bids
			(
				id uuid NOT NULL DEFAULT uuid_generate_v4(),
				created_at timestamp with time zone,
				updated_at timestamp with time zone,
				deleted_at timestamp with time zone,
				name character varying(100) COLLATE pg_catalog."default" NOT NULL,
				description character varying(500) COLLATE pg_catalog."default",
				status bid_status NOT NULL,
				tender_id uuid,
				employee_username text COLLATE pg_catalog."default" NOT NULL,
				organization_id uuid NOT NULL,
				author_type bid_author_type NOT NULL,
				version bigint DEFAULT 1,
				CONSTRAINT bids_pkey PRIMARY KEY (id),
				CONSTRAINT fk_bids_organization FOREIGN KEY (organization_id)
					REFERENCES organization (id) MATCH SIMPLE
					ON UPDATE NO ACTION
					ON DELETE NO ACTION,
				CONSTRAINT fk_bids_tender FOREIGN KEY (tender_id)
					REFERENCES tenders (id) MATCH SIMPLE
					ON UPDATE NO ACTION
					ON DELETE NO ACTION
			);
			
			CREATE INDEX IF NOT EXISTS idx_bids_deleted_at
				ON bids USING btree
				(deleted_at ASC NULLS LAST)
`)

	db.Exec(`
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'organization_type') THEN
								CREATE TYPE organization_type AS ENUM (
									'IE',
									'LLC',
									'JSC'
								);
				END IF;
			END $$;
	`)

	db.Exec(`
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tender_service_type') THEN
								CREATE TYPE tender_service_type AS ENUM (
									'Construction',
									'Delivery',
									'Manufacture'
									);
				END IF;
			END $$;
	`)

	db.Exec(`
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tender_status') THEN
								CREATE TYPE tender_status AS ENUM (
										'Created',
										'Published',
										'Closed'
										);
				END IF;
			END $$;
	`)

	db.Exec(`
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'bid_status') THEN
							CREATE TYPE bid_status AS ENUM (
									'Created',
									'Published',
									'Canceled',
									'Approved',
									'Rejected'
									);
				END IF;
			END $$;
	`)

	db.Exec(`
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'bid_author_type') THEN
								CREATE TYPE bid_author_type AS ENUM (
										'Organization',
										'User'
										);
				END IF;
			END $$;
	`)

	db.Exec(`
	CREATE OR REPLACE FUNCTION uuid_generate_v1(
		)
		RETURNS uuid
		LANGUAGE 'c'
		COST 1
		VOLATILE STRICT PARALLEL SAFE 
	AS '$libdir/uuid-ossp', 'uuid_generate_v1'
	;

	CREATE OR REPLACE FUNCTION uuid_generate_v1mc(
		)
		RETURNS uuid
		LANGUAGE 'c'
		COST 1
		VOLATILE STRICT PARALLEL SAFE 
	AS '$libdir/uuid-ossp', 'uuid_generate_v1mc'
	;
	
	CREATE OR REPLACE FUNCTION uuid_generate_v3(
		namespace uuid,
		name text)
		RETURNS uuid
		LANGUAGE 'c'
		COST 1
		IMMUTABLE STRICT PARALLEL SAFE 
	AS '$libdir/uuid-ossp', 'uuid_generate_v3'
	;
	
	CREATE OR REPLACE FUNCTION uuid_generate_v4(
		)
		RETURNS uuid
		LANGUAGE 'c'
		COST 1
		VOLATILE STRICT PARALLEL SAFE 
	AS '$libdir/uuid-ossp', 'uuid_generate_v4'
	;
	
	CREATE OR REPLACE FUNCTION uuid_generate_v5(
		namespace uuid,
		name text)
		RETURNS uuid
		LANGUAGE 'c'
		COST 1
		IMMUTABLE STRICT PARALLEL SAFE 
	AS '$libdir/uuid-ossp', 'uuid_generate_v5'
	;
	
	CREATE OR REPLACE FUNCTION uuid_nil(
		)
		RETURNS uuid
		LANGUAGE 'c'
		COST 1
		IMMUTABLE STRICT PARALLEL SAFE 
	AS '$libdir/uuid-ossp', 'uuid_nil'
	;
	
	CREATE OR REPLACE FUNCTION uuid_ns_dns(
		)
		RETURNS uuid
		LANGUAGE 'c'
		COST 1
		IMMUTABLE STRICT PARALLEL SAFE 
	AS '$libdir/uuid-ossp', 'uuid_ns_dns'
	;
	
	CREATE OR REPLACE FUNCTION uuid_ns_oid(
		)
		RETURNS uuid
		LANGUAGE 'c'
		COST 1
		IMMUTABLE STRICT PARALLEL SAFE 
	AS '$libdir/uuid-ossp', 'uuid_ns_oid'
	;
	
	
	CREATE OR REPLACE FUNCTION uuid_ns_url(
		)
		RETURNS uuid
		LANGUAGE 'c'
		COST 1
		IMMUTABLE STRICT PARALLEL SAFE 
	AS '$libdir/uuid-ossp', 'uuid_ns_url'
	;
	
	CREATE OR REPLACE FUNCTION uuid_ns_x500(
		)
		RETURNS uuid
		LANGUAGE 'c'
		COST 1
		IMMUTABLE STRICT PARALLEL SAFE 
	AS '$libdir/uuid-ossp', 'uuid_ns_x500'
	;
`)
}

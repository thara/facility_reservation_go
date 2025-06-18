# Facility Reservation System Documentation

Welcome to the documentation for the Facility Reservation System - a REST API service for managing facility bookings and user administration.

## 📋 Documentation Index

### 📋 Product
- **[Product Requirements Document (PRD)](prd.md)** - Complete product vision, user needs, and implementation roadmap
  - User personas, workflows, and success metrics
  - Functional requirements organized by implementation phases
  - Competitive analysis and market positioning

### 🏗️ Architecture
- **[Architecture Overview](architecture.md)** - Complete system architecture using arc42 template
  - System context, building blocks, runtime views
  - Quality requirements, deployment, and crosscutting concepts
  - Links to all related ADRs and design documents

### 🎯 Architecture Decision Records (ADRs)
- **[ADR Index](adr/)** - Key architectural decisions and their rationale
  - [ADR-001: Use TypeSpec over OpenAPI YAML](adr/0001-use-typespec-over-openapi-yaml.md)
  - [ADR-002: Use sqlc over ORM](adr/0002-use-sqlc-over-orm.md)
  - [ADR-003: Avoid Database Triggers](adr/0003-avoid-database-triggers.md)

### 📐 Design Documents
- **[Design Docs Index](design-docs/)** - Detailed designs for complex systems
  - [Code Generation Pipeline](design-docs/001-code-generation-pipeline.md)
  - [API Contract Ecosystem](design-docs/002-api-contract-ecosystem.md)
  - [Development Workflow](design-docs/003-development-workflow.md)


## 🚀 Quick Start

### For Product Managers & Stakeholders
1. **Start here**: [Product Requirements Document (PRD)](prd.md) - Complete product vision and roadmap
2. **Current status**: Phase 0 (foundation) complete, Phase 1 (core booking) in detailed planning
3. **Success metrics**: User adoption, booking efficiency, and system reliability targets

### For New Developers
1. **Product context**: [PRD](prd.md) - Understand what we're building and why
2. **System design**: [Architecture Overview](architecture.md) - Technical system architecture
3. **Daily workflow**: [Development Workflow](design-docs/003-development-workflow.md) for productive development

### For Architects
1. **Product strategy**: [PRD](prd.md) - Business requirements and success criteria
2. **System overview**: [Architecture Overview](architecture.md) - Complete arc42 documentation
3. **Decision context**: [ADR Index](adr/) - Rationale behind architectural choices
4. **Complex systems**: [Design Docs](design-docs/) - Implementation details for sophisticated workflows

### For API Consumers
1. **API specification**: `spec/main.tsp` - TypeSpec definition of all endpoints
2. **Client generation**: [API Contract Ecosystem](design-docs/002-api-contract-ecosystem.md) - How to generate client SDKs
3. **OpenAPI schema**: `spec/tsp-output/schema/3.1.0/openapi.yaml` - Generated OpenAPI specification

## 🛠️ Development Quick Reference

For complete development workflow, see [Development Workflow](design-docs/003-development-workflow.md).

### Key Files
- `spec/main.tsp` - API specification (TypeSpec)
- `_db/schema.sql` - Database schema (source of truth)
- `_db/query_*.sql` - SQL queries for code generation
- `internal/api_service.go` - Business logic implementation

## 📚 Documentation Principles

This documentation follows a structured approach with clear ownership:
- **ADRs**: Document architectural decisions with rationale
- **Design Docs**: Detailed implementation patterns for complex systems  
- **Architecture**: Comprehensive system overview using arc42 template

For contributing guidelines, see individual README files in [adr/](adr/) and [design-docs/](design-docs/) directories.

## 🎯 System Overview

The Facility Reservation System is built with a **schema-driven development** approach using TypeSpec for API specification, sqlc for database code generation, and Atlas for schema management.

For complete system architecture and design decisions, see [Architecture Documentation](architecture.md).

## 🏷️ Technology Stack

For detailed technology choices and architecture, see [Architecture Documentation](architecture.md).

---

*For detailed information about any topic, follow the links above or browse the documentation directories.*

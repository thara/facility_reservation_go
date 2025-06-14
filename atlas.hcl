// Atlas configuration for facility reservation API
data "external_schema" "app" {
  program = [
    "sh", "-c", "cat schema/schema.sql"
  ]
}

env "dev" {
  src = data.external_schema.app.url
  dev = "postgres://postgres:postgres@localhost:5432/facility_reservation_dev?sslmode=disable"
  url = "postgres://postgres:postgres@localhost:5432/facility_reservation_dev?sslmode=disable"
  
  migration {
    dir = "file://migrations"
  }
  
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

env "test" {
  src = data.external_schema.app.url
  dev = "postgres://postgres:postgres@localhost:5433/facility_reservation_test?sslmode=disable"
  url = "postgres://postgres:postgres@localhost:5433/facility_reservation_test?sslmode=disable"
  
  migration {
    dir = "file://migrations"
  }
}
package db

import (
    "context"
    "strconv"

    "github.com/jmoiron/sqlx"
    _ "github.com/jackc/pgx/v5/stdlib"
    "transjakarta-fleet/internal/models"
)

type DB struct{ *sqlx.DB }

func New(dsn string) (*DB, error) {
    db, err := sqlx.Connect("pgx", dsn)
    if err != nil { return nil, err }
    return &DB{db}, nil
}

func (d *DB) Close() error { return d.DB.Close() }

func (d *DB) InsertLocation(ctx context.Context, m models.LocationMessage) error {
    _, err := d.ExecContext(ctx, `INSERT INTO vehicle_locations(vehicle_id, latitude, longitude, ts) VALUES($1,$2,$3,$4)`, m.VehicleID, m.Latitude, m.Longitude, m.Timestamp)
    return err
}

func (d *DB) GetLastLocation(ctx context.Context, vehicleID string) (*models.LocationMessage, error) {
    var out models.LocationMessage
    q := `SELECT vehicle_id, latitude, longitude, ts FROM vehicle_locations WHERE vehicle_id=$1 ORDER BY ts DESC LIMIT 1`
    if err := d.GetContext(ctx, &out, q, vehicleID); err != nil { return nil, err }
    return &out, nil
}

func (d *DB) GetHistory(ctx context.Context, vehicleID string, hasStart, hasEnd bool, start, end int64) ([]models.LocationMessage, error) {
    var rows []models.LocationMessage
    var args []interface{}
    q := `SELECT vehicle_id, latitude, longitude, ts FROM vehicle_locations WHERE vehicle_id=$1`
    args = append(args, vehicleID)
    argCount := 1
    if hasStart {
        argCount++
        q += ` AND ts >= $` + strconv.Itoa(argCount)
        args = append(args, start)
    }
    if hasEnd {
        argCount++
        q += ` AND ts <= $` + strconv.Itoa(argCount)
        args = append(args, end)
    }
    q += ` ORDER BY ts ASC`
    if err := d.SelectContext(ctx, &rows, q, args...); err != nil { return nil, err }
    return rows, nil
}

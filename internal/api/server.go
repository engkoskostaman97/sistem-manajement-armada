package api

import (
    "database/sql"
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "transjakarta-fleet/internal/db"
)

func NewServer(dbconn *db.DB) *gin.Engine {
    r := gin.Default()

    r.GET("/vehicles/:vehicle_id/location", func(c *gin.Context) {
        id := c.Param("vehicle_id")
        loc, err := dbconn.GetLastLocation(c.Request.Context(), id)
        if err != nil {
            if err == sql.ErrNoRows {
                c.JSON(http.StatusNotFound, gin.H{"error": "no location data found for vehicle"})
                return
            }
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, loc)
    })

    r.GET("/vehicles/:vehicle_id/history", func(c *gin.Context) {
        id := c.Param("vehicle_id")
        startS := c.Query("start")
        endS := c.Query("end")
        var start, end int64
        var err error
        if startS != "" {
            start, err = strconv.ParseInt(startS, 10, 64)
            if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start parameter"})
                return
            }
        }
        if endS != "" {
            end, err = strconv.ParseInt(endS, 10, 64)
            if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end parameter"})
                return
            }
        }
        rows, err := dbconn.GetHistory(c.Request.Context(), id, startS != "", endS != "", start, end)
        if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
        c.JSON(http.StatusOK, rows)
    })

    return r
}

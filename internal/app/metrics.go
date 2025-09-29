package app

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
)

//var (
//	tableRowCount = prometheus.NewGaugeVec(
//		prometheus.GaugeOpts{
//			Name: "db_table_row_count",
//			Help: "Number of rows per table",
//		},
//		[]string{"table"},
//	)
//
//	tableInsertCount = prometheus.NewCounterVec(
//		prometheus.CounterOpts{
//			Name: "db_table_insert_total",
//			Help: "Total inserts per table",
//		},
//		[]string{"table"},
//	)
//)
//
//func init() {
//	prometheus.MustRegister(tableRowCount)
//	prometheus.MustRegister(tableInsertCount)
//}
//
//func startTableMetricsCollector(ctx context.Context, db *pgxpool.Pool, logger *zap.Logger) {
//
//	ticker := time.NewTicker(10 * time.Second)
//	defer ticker.Stop()
//	lastRowCounts := map[string]int64{}
//
//	tables := []string{"author", "book", "author_book"}
//	for {
//		select {
//		case <-ctx.Done():
//			return
//		case <-ticker.C:
//			for _, table := range tables {
//				var count int64
//				err := db.QueryRow(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count)
//				if err != nil {
//					logger.Error("Failed to count rows.", zap.String("table", table), zap.Error(err))
//					continue
//				}
//				tableRowCount.WithLabelValues(table).Set(float64(count))
//
//				last, ok := lastRowCounts[table]
//				if ok && count > last {
//					tableInsertCount.WithLabelValues(table).Add(float64(count - last))
//				}
//				lastRowCounts[table] = count
//			}
//		}
//	}
//}

func runMetricsServer(logger *zap.Logger, port string) {
	logger.Info("Starting metrics server.", zap.String("port", port))
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Fatal("Can not start metrics server.", zap.Error(err))
	}
}
